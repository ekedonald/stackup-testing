pipeline {
    agent any
    
    environment {
        DOCKER_IMAGE = 'delivery-tracker'
        DOCKER_TAG = "${BUILD_NUMBER}"
        REMOTE_DIR = '/home/ubuntu/stackup-testing'
        GIT_REPO = 'https://github.com/ekedonald/stackup-testing.git'
    }
    
    stages {
        stage('Create .env and Build Image') {
            steps {
                script {
                    withCredentials([file(credentialsId: 'env-file-secrets', variable: 'ENV_FILE')]) {
                        sh '''
                            cp "$ENV_FILE" .env
                            git clone $GIT_REPO
                            cd stackup-testing
                            docker build -t $DOCKER_IMAGE:$DOCKER_TAG .
                            docker save $DOCKER_IMAGE:$DOCKER_TAG > ../image.tar
                            cd ..
                            rm -rf stackup-testing
                        '''
                    }
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
                                        sourceFiles: "image.tar,.env",
                                        remoteDirectory: "",
                                        execCommand: """
                                            # Load Docker image
                                            docker load < /tmp/image.tar
                                            rm /tmp/image.tar
                                            
                                            # Setup application
                                            cd /home/ubuntu
                                            rm -rf ${REMOTE_DIR}
                                            git clone ${GIT_REPO}
                                            mv /tmp/.env ${REMOTE_DIR}/
                                            
                                            # Deploy
                                            cd ${REMOTE_DIR}
                                            sed -i "s|build: .|image: ${DOCKER_IMAGE}:${DOCKER_TAG}|g" compose.yaml
                                            docker compose up -d
                                        """
                                    )
                                ]
                            )
                        ]
                    )
                }
            }
        }
    }
    
    post {
        always {
            sh '''
                rm -f image.tar
                rm -f .env
            '''
            cleanWs()
        }
    }
}