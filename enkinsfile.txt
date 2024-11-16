pipeline {
    agent any
    
    environment {
        DOCKER_IMAGE = 'delivery-tracker'
        DOCKER_TAG = "${BUILD_NUMBER}"
        REMOTE_USER = 'ubuntu'
        REMOTE_HOST = '52.90.24.129'
        REMOTE_DIR = '/home/ubuntu/stackup-testing'  // Updated to direct stackup directory
        PEM_PATH = "/tmp/deploy-key-${BUILD_NUMBER}.pem"
    }
    
    stages {
        stage('Setup SSH') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(credentialsId: 'ssh-pem-key', 
                                                     keyFileVariable: 'SSH_KEY')]) {
                        sh """
                            cp "\$SSH_KEY" ${PEM_PATH}
                            chmod 600 ${PEM_PATH}
                            ssh-keyscan -H ${REMOTE_HOST} >> /tmp/known_hosts_${BUILD_NUMBER}
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
        
        stage('Transfer Files to Remote') {
            steps {
                script {
                    sh """
                        scp -i ${PEM_PATH} -o UserKnownHostsFile=/tmp/known_hosts_${BUILD_NUMBER} \
                            ${DOCKER_IMAGE}.tar ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/
                        scp -i ${PEM_PATH} -o UserKnownHostsFile=/tmp/known_hosts_${BUILD_NUMBER} \
                            .env ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/
                    """
                }
            }
        }
        
        stage('Deploy on Remote Server') {
            steps {
                script {
                    sh """
                        ssh -i ${PEM_PATH} -o UserKnownHostsFile=/tmp/known_hosts_${BUILD_NUMBER} \
                            ${REMOTE_USER}@${REMOTE_HOST} '\
                            cd ${REMOTE_DIR} && \
                            docker load < ${DOCKER_IMAGE}.tar && \
                            rm ${DOCKER_IMAGE}.tar && \
                            sed -i "s|build: .|image: ${DOCKER_IMAGE}:${DOCKER_TAG}|g" compose.yaml && \
                            docker compose up -d'
                    """
                }
            }
        }
    }
    
    post {
        always {
            // Clean up all temporary files
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