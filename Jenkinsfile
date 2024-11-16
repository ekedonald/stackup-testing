pipeline {
    agent any
    
    environment {
        DOCKER_IMAGE = 'delivery-tracker'
        DOCKER_TAG = "${BUILD_NUMBER}"
        PEM_PATH = "/tmp/deploy-key-${BUILD_NUMBER}.pem"
        TEMP_DIR = "/tmp/deployment-${BUILD_NUMBER}"
        GIT_REPO = 'https://github.com/ekedonald/stackup-testing.git'
    }
    
    stages {
        stage('Setup SSH') {
            steps {
                script {
                    withCredentials([
                        string(credentialsId: 'remote-user', variable: 'REMOTE_USER'),
                        string(credentialsId: 'remote-host', variable: 'REMOTE_HOST'),
                        string(credentialsId: 'remote-dir', variable: 'REMOTE_DIR'),
                        sshUserPrivateKey(credentialsId: 'ssh-pem-key', keyFileVariable: 'SSH_KEY')
                    ]) {
                        sh """
                            cp "\$SSH_KEY" ${PEM_PATH}
                            chmod 600 ${PEM_PATH}
                            ssh-keyscan -H \$REMOTE_HOST >> /tmp/known_hosts_${BUILD_NUMBER}
                        """
                    }
                }
            }
        }
        
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
        
        stage('Transfer and Deploy') {
            steps {
                script {
                    withCredentials([
                        string(credentialsId: 'remote-user', variable: 'REMOTE_USER'),
                        string(credentialsId: 'remote-host', variable: 'REMOTE_HOST'),
                        string(credentialsId: 'remote-dir', variable: 'REMOTE_DIR')
                    ]) {
                        // Create temp directory and transfer files
                        sh """
                            ssh -i ${PEM_PATH} -o UserKnownHostsFile=/tmp/known_hosts_${BUILD_NUMBER} \
                                \$REMOTE_USER@\$REMOTE_HOST '\
                                mkdir -p ${TEMP_DIR}'

                            scp -i ${PEM_PATH} -o UserKnownHostsFile=/tmp/known_hosts_${BUILD_NUMBER} \
                                ${DOCKER_IMAGE}.tar \$REMOTE_USER@\$REMOTE_HOST:${TEMP_DIR}/
                            scp -i ${PEM_PATH} -o UserKnownHostsFile=/tmp/known_hosts_${BUILD_NUMBER} \
                                .env \$REMOTE_USER@\$REMOTE_HOST:${TEMP_DIR}/

                            # Execute deployment commands on remote server
                            ssh -i ${PEM_PATH} -o UserKnownHostsFile=/tmp/known_hosts_${BUILD_NUMBER} \
                                \$REMOTE_USER@\$REMOTE_HOST '\
                                git clone ${GIT_REPO} \$REMOTE_DIR && \
                                docker load < ${TEMP_DIR}/${DOCKER_IMAGE}.tar && \
                                rm -rf ${TEMP_DIR} && \
                                cd \$REMOTE_DIR && \
                                cp ${TEMP_DIR}/.env . && \
                                sed -i "s|build: .|image: ${DOCKER_IMAGE}:${DOCKER_TAG}|g" compose.yaml && \
                                docker compose up -d'
                        """
                    }
                }
            }
        }
    }
    
    post {
        always {
            sh """
                rm -f ${PEM_PATH}
                rm -f /tmp/known_hosts_${BUILD_NUMBER}
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