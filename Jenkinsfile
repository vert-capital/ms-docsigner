@NonCPS
def cancelPreviousBuilds() {
  String jobname = env.JOB_NAME
  int currentBuildNum = env.BUILD_NUMBER.toInteger()

  def job = Jenkins.instance.getItemByFullName(jobname)
  for (build in job.builds) {
    if (build.isBuilding() && currentBuildNum > build.getNumber().toInteger()) {
      build.doStop();
      echo "Build ${build.getNumber()} cancelled"
    }
  }
}

pipeline {
    environment {
        registry = "197272534240.dkr.ecr.us-east-1.amazonaws.com"
        registryCredential = "ecr:us-east-1:aws_vert"
        dockerImageName = ""
    }

    agent {
        docker {
            image "akaytatsu/cibuilder:latest"
        }
    }

    stages {

        stage('Init') {
            steps {
                cancelPreviousBuilds()
            }
        }

        stage('Code Checkout') {
            steps {
                checkout scm
            }
        }

        // Stages comuns (todos os branches passam por aqui)
        stage('Build Docker Images') {
            steps {
                script {
                    sh 'cp -f src/.env.sample src/.env'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml build'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml up -d --no-build'
                }
            }
        }

        stage('stop containers') {
            steps {
                script {
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
                }
            }
        }
                                                       
        stage('build Container Register Staging') {
            when {
                expression {
                    return env.GIT_BRANCH == 'develop'
                }
            }
        
            steps {
                script {
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImageName = "ms-docsigner-stg"
                        dockerImage = docker.build(dockerImageName, "./src")
                        dockerImage.push("$BUILD_NUMBER")
                        dockerImage.push("latest")
                    }
                }
        
                script{
                    sh "docker rmi $registry/$dockerImageName:$BUILD_NUMBER || true"
                    sh "docker rmi $registry/$dockerImageName:latest || true"
                }
            }
        }

                                                       
        stage('build Container Register Homologation') {
            when {
                expression {
                    return env.GIT_BRANCH == 'homolog'
                }
            }

            steps {
                script {
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImageName = "ms-docsigner-hml"
                        dockerImage = docker.build(dockerImageName, "./src")
                         dockerImage.push("$BUILD_NUMBER")
                        dockerImage.push("latest")
                    }
                }

                script {
                    sh "docker rmi ${registry}/${dockerImageName}:${BUILD_NUMBER} || docker rmi ${dockerImageName}:${BUILD_NUMBER} || true"
                    sh "docker rmi ${registry}/${dockerImageName}:latest || docker rmi ${dockerImageName}:latest || true"
                }
            }
        }


        stage('build Container Register Production') {
            when {
                expression {
                    return env.GIT_BRANCH == 'master'
                }
            }

            steps {
                script {
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImageName = "ms-docsigner-prd"
                        dockerImage = docker.build(dockerImageName, "./src")
                        dockerImage.push("${BUILD_NUMBER}")
                        dockerImage.push("latest")
                    }
                }

                script {
                    sh "docker rmi ${registry}/${dockerImageName}:${BUILD_NUMBER} || docker rmi ${dockerImageName}:${BUILD_NUMBER} || true"
                    sh "docker rmi ${registry}/${dockerImageName}:latest || docker rmi ${dockerImageName}:latest || true"
                }
            }
        }


        stage('Deploy to Staging Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'develop'
                }
            }

            steps {
                script {
                    withCredentials([string(credentialsId: "ARGOCD_SERVER", variable: 'ARGOCD_SERVER')]) {
                        withCredentials([string(credentialsId: "argocd-homolog", variable: 'ARGOCD_AUTH_TOKEN')]) {
                            sh "argocd --grpc-web app actions run ms-docsigner-stg restart --kind Deployment --all"
                        }
                    }
                }
            }
        }

 
        stage('Deploy to Homolog Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'homolog'
                }
            }

            steps {
                script {
                    withCredentials([string(credentialsId: "ARGOCD_SERVER", variable: 'ARGOCD_SERVER')]) {
                        withCredentials([string(credentialsId: "argocd-homolog", variable: 'ARGOCD_AUTH_TOKEN')]) {
                            sh "argocd --grpc-web app actions run ms-docsigner-hml restart --kind Deployment --all"
                        }
                    }
                }
            }
        }

    
        stage('Deploy to Production Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'prd'
                }
            }

            steps {
                script {
                    withCredentials([string(credentialsId: "ARGOCD_SERVER", variable: 'ARGOCD_SERVER')]) {
                        withCredentials([string(credentialsId: "argocd-production", variable: 'ARGOCD_AUTH_TOKEN')]) {
                            sh "argocd --grpc-web app actions run ms-docsigner-prd restart --kind Deployment --all"
                        }
                    }
                }
            }
        }

    }

    post {
        always {
            echo "Stop Docker image"
            script{
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
            }
        }

        success {
            echo "Notify bitbucket success"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
            }
        }

        failure {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
            }
        }

        aborted {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
            }
        }

        unsuccessful {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
            }
        }

    }
}
