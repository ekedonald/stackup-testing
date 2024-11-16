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
                        sh 'cp $ENV_FILE .env'
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
                    sshPublisher(
                        publishers: [
                            sshPublisherDesc(
                                configName: 'remote-server',
                                transfers: [
                                    sshTransfer(
                                        sourceFiles: "${DOCKER_IMAGE}.tar,.env",
                                        removePrefix: '',
                                        remoteDirectory: "tmp",
                                        execCommand: """
                                            # Create temporary directory
                                            mkdir -p ${TMP_DIR}
                                            
                                            # Move files from default SSH transfer directory to our temp directory
                                            mv /tmp/${DOCKER_IMAGE}.tar ${TMP_DIR}/
                                            mv /tmp/.env ${TMP_DIR}/
                                            
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
                                        remoteDirectory: "${REMOTE_DIR}",
                                        execCommand: """
                                            cd ${REMOTE_DIR}
                                            
                                            # Move .env file from temp directory to final location
                                            mv ${TMP_DIR}/.env ${REMOTE_DIR}/
                                            
                                            # Clean up temp directory
                                            rm -rf ${TMP_DIR}
                                            
                                            sed -i "s|build: .|image: ${DOCKER_IMAGE}:${DOCKER_TAG}|g" compose.yaml
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