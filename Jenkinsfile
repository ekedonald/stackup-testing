pipeline {
    agent any
    
    environment {
        DOCKER_IMAGE = 'delivery-tracker'
        DOCKER_TAG = "${BUILD_NUMBER}"
        REMOTE_DIR = '/home/ubuntu/stackup-testing'
        TMP_DIR = '/tmp/deployment-${BUILD_NUMBER}'
        GIT_REPO = 'https://github.com/ekedonald/stackup-testing.git'
    }
    
    stages {
        stage('Create .env File') {
            steps {
                script {
                    withCredentials([file(credentialsId: 'env-file-secrets', variable: 'ENV_FILE')]) {
                        sh '''
                            cp "$ENV_FILE" .env
                            ls -la .env  # Debug: verify file exists and permissions
                        '''
                    }
                }
            }
        }
        
        stage('Build Docker Image') {
            steps {
                script {
                    sh "docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} ."
                    sh "docker save ${DOCKER_IMAGE}:${DOCKER_TAG} > ${DOCKER_IMAGE}.tar"
                }
            }
        }
        
        stage('Transfer Files') {
            steps {
                script {
                    // Debug: List files before transfer
                    sh 'ls -la ${DOCKER_IMAGE}.tar .env'
                    
                    sshPublisher(
                        publishers: [
                            sshPublisherDesc(
                                configName: 'remote-server',
                                transfers: [
                                    sshTransfer(
                                        sourceFiles: "${DOCKER_IMAGE}.tar,.env",
                                        removePrefix: '',
                                        remoteDirectory: "",
                                        execCommand: """
                                            # Debug: List files in /tmp after transfer
                                            echo "Files in /tmp:"
                                            ls -la /tmp/.env /tmp/${DOCKER_IMAGE}.tar || echo "Files not found in /tmp"
                                            
                                            # Create temporary directory
                                            mkdir -p ${TMP_DIR}
                                            echo "Created directory ${TMP_DIR}"
                                            
                                            # Debug: List current directory
                                            pwd
                                            ls -la
                                            
                                            # Move files to temp directory with error checking and verbose output
                                            for file in "/tmp/${DOCKER_IMAGE}.tar" "/tmp/.env"; do
                                                if [ -f "\$file" ]; then
                                                    echo "Moving \$file to ${TMP_DIR}/"
                                                    mv "\$file" "${TMP_DIR}/"
                                                else
                                                    echo "File \$file not found!"
                                                    exit 1
                                                fi
                                            done
                                            
                                            # Debug: List files in temp directory
                                            echo "Files in ${TMP_DIR}:"
                                            ls -la ${TMP_DIR}
                                            
                                            # Clone or update git repository
                                            if [ ! -d "${REMOTE_DIR}" ]; then
                                                git clone ${GIT_REPO} ${REMOTE_DIR}
                                            else
                                                cd ${REMOTE_DIR}
                                                git fetch origin
                                                git reset --hard origin/main
                                            fi
                                            
                                            # Load the Docker image
                                            docker load < ${TMP_DIR}/${DOCKER_IMAGE}.tar
                                            rm ${TMP_DIR}/${DOCKER_IMAGE}.tar
                                        """
                                    )
                                ],
                                verbose: true
                            )
                        ]
                    )
                }
            }
        }

        stage('Deploy') {
            steps {
                script {
                    sshPublisher(
                        publishers: [
                            sshPublisherDesc(
                                configName: 'remote-server',
                                transfers: [
                                    sshTransfer(
                                        sourceFiles: '',
                                        removePrefix: '',
                                        remoteDirectory: '',
                                        execCommand: """
                                            # Debug: Check temp directory contents
                                            echo "Contents of ${TMP_DIR}:"
                                            ls -la ${TMP_DIR} || echo "Directory not found"
                                            
                                            # Verify .env exists in temp directory
                                            if [ ! -f "${TMP_DIR}/.env" ]; then
                                                echo ".env file not found in ${TMP_DIR}!"
                                                exit 1
                                            fi
                                            
                                            cd ${REMOTE_DIR}
                                            
                                            # Move .env file with error checking
                                            echo "Moving .env file to ${REMOTE_DIR}"
                                            mv "${TMP_DIR}/.env" "${REMOTE_DIR}/"
                                            
                                            # Verify the move was successful
                                            if [ -f ".env" ]; then
                                                echo ".env file successfully moved to ${REMOTE_DIR}"
                                            else
                                                echo "Failed to move .env file to ${REMOTE_DIR}"
                                                exit 1
                                            fi
                                            
                                            # Clean up temp directory
                                            rm -rf ${TMP_DIR}
                                            
                                            sed -i "s|build: .|image: ${DOCKER_IMAGE}:${DOCKER_TAG}|g" compose.yaml
                                            
                                            # Debug: Verify .env contents (without showing sensitive data)
                                            echo "Checking .env file:"
                                            if [ -s ".env" ]; then
                                                echo ".env file exists and is not empty"
                                            else
                                                echo ".env file is empty or missing"
                                                exit 1
                                            fi
                                            
                                            docker compose up -d
                                        """
                                    )
                                ],
                                verbose: true
                            )
                        ]
                    )
                }
            }
        }
    }
    
    post {
        always {
            sh """
                rm -f ${DOCKER_IMAGE}.tar
                rm -f .env
            """
            cleanWs()
        }
        success {
            echo 'Deployment completed successfully!'
        }
        failure {
            echo 'Deployment failed!'
        }
    }
}