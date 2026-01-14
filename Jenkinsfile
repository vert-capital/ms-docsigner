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
        // removi dockerImageName daqui para evitar campo global mutável
    }

    agent {
        docker {
            image "akaytatsu/cibuilder:latest"
        }
    }

    stages {

        stage('Init') {
            steps {
                script {
                    cancelPreviousBuilds()

                    // Normaliza a variável para comparações exatas funcionarem
                    def rawBranch = env.GIT_BRANCH ?: env.BRANCH_NAME ?: ""
                    rawBranch = rawBranch.toString()
                    rawBranch = rawBranch.replaceFirst(/^origin\//, "")
                    rawBranch = rawBranch.replaceFirst(/^refs\\/heads\\//, "")
                    env.GIT_BRANCH = rawBranch
                    echo "Normalized GIT_BRANCH -> ${env.GIT_BRANCH}"
                }
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

        // Build & push Staging (develop)
        stage('build Container Register Staging') {
            when {
                expression {
                    return env.GIT_BRANCH == 'develop'
                }
            }

            steps {
                script {
                    def dockerImageName = "ms-docsigner-stg"
                    def dockerImage
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImage = docker.build(dockerImageName, "./src")
                        dockerImage.push("${BUILD_NUMBER}")
                        dockerImage.push("latest")
                    }
                    // cleanup
                    sh "docker rmi ${registry}/${dockerImageName}:${BUILD_NUMBER} || docker rmi ${dockerImageName}:${BUILD_NUMBER} || true"
                    sh "docker rmi ${registry}/${dockerImageName}:latest || docker rmi ${dockerImageName}:latest || true"
                }
            }
        }

        // Build & push Homolog (homolog)
        stage('build Container Register Homologation') {
            when {
                expression {
                    return env.GIT_BRANCH == 'homolog'
                }
            }

            steps {
                script {
                    def dockerImageName = "ms-docsigner-hml"
                    def dockerImage
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImage = docker.build(dockerImageName, "./src")
                        dockerImage.push("${BUILD_NUMBER}")
                        dockerImage.push("latest")
                    }
                    sh "docker rmi ${registry}/${dockerImageName}:${BUILD_NUMBER} || docker rmi ${dockerImageName}:${BUILD_NUMBER} || true"
                    sh "docker rmi ${registry}/${dockerImageName}:latest || docker rmi ${dockerImageName}:latest || true"
                }
            }
        }

        // Build & push Production (master)
        stage('build Container Register Production') {
            when {
                expression {
                    return env.GIT_BRANCH == 'master'
                }
            }

            steps {
                script {
                    def dockerImageName = "ms-docsigner-prd"
                    def dockerImage
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImage = docker.build(dockerImageName, "./src")
                        dockerImage.push("${BUILD_NUMBER}")
                        dockerImage.push("latest")
                    }
                    sh "docker rmi ${registry}/${dockerImageName}:${BUILD_NUMBER} || docker rmi ${dockerImageName}:${BUILD_NUMBER} || true"
                    sh "docker rmi ${registry}/${dockerImageName}:latest || docker rmi ${dockerImageName}:latest || true"
                }
            }
        }

        // Deploy to Staging Environment (develop)
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

        // Deploy to Homolog Environment (homolog)
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

        // Deploy to Production Environment (master)
        stage('Deploy to Production Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'master'
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
