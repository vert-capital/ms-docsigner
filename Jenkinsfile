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


def buildAndPushDockerImage(String imageName) {
    docker.withRegistry("https://$registry", registryCredential) {
        def dockerImage = docker.build(imageName, "./src")
        dockerImage.push("${BUILD_NUMBER}")
        dockerImage.push("latest")
    }
    sh "docker rmi ${registry}/${imageName}:${BUILD_NUMBER} || docker rmi ${imageName}:${BUILD_NUMBER} || true"
    sh "docker rmi ${registry}/${imageName}:latest || docker rmi ${imageName}:latest || true"
}


def deployToApp(String appName, String credId) {
    withCredentials([string(credentialsId: "ARGOCD_SERVER", variable: 'ARGOCD_SERVER')]) {
        withCredentials([string(credentialsId: credId, variable: 'ARGOCD_AUTH_TOKEN')]) {
            sh "argocd --grpc-web app actions run ${appName} restart --kind Deployment --all"
        }
    }
}

pipeline {
    environment {
        registry = "197272534240.dkr.ecr.us-east-1.amazonaws.com"
        registryCredential = "ecr:us-east-1:aws_vert"

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


                    def rawBranch = env.BRANCH_NAME ?: env.GIT_BRANCH ?: ""
                    rawBranch = rawBranch.toString()
                    rawBranch = rawBranch.replaceFirst(/^origin\//, "")
                    rawBranch = rawBranch.replaceFirst(/^refs\\/heads\\//, "")
                    env.GIT_BRANCH = rawBranch
                    if (!env.BRANCH_NAME) { env.BRANCH_NAME = rawBranch }
                    echo "Normalized branch -> BRANCH_NAME='${env.BRANCH_NAME}' GIT_BRANCH='${env.GIT_BRANCH}'"
                }
            }
        }

        stage('Code Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Docker Images (for tests)') {
            steps {
                script {

                    sh 'cp -f src/.env.sample src/.env || true'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml build'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml up -d --no-build'
                }
            }
        }


        stage('Stop test containers') {
            steps {
                script {
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
                }
            }
        }

        stage('Build & Push Image (per-branch)') {
            when {
                expression { return env.GIT_BRANCH == 'develop' || env.GIT_BRANCH == 'homolog' || env.GIT_BRANCH == 'master' }
            }
            steps {
                script {
                    if (env.GIT_BRANCH == 'develop') {
                        buildAndPushDockerImage("ms-docsigner-stg")
                    } else if (env.GIT_BRANCH == 'homolog') {
                        buildAndPushDockerImage("ms-docsigner-hml")
                    } else if (env.GIT_BRANCH == 'master') {
                        buildAndPushDockerImage("ms-docsigner-prd")
                    } else {
                        echo "No image push for branch ${env.GIT_BRANCH}"
                    }
                }
            }
        }

        stage('Deploy (per-branch)') {
            when {
                expression { return env.GIT_BRANCH == 'develop' || env.GIT_BRANCH == 'homolog' || env.GIT_BRANCH == 'master' }
            }
            steps {
                script {
                    if (env.GIT_BRANCH == 'develop') {
                        deployToApp('ms-docsigner-stg', 'argocd-homolog')
                    } else if (env.GIT_BRANCH == 'homolog') {
                        deployToApp('ms-docsigner-hml', 'argocd-homolog')
                    } else if (env.GIT_BRANCH == 'master') {
                        deployToApp('ms-docsigner-prd', 'argocd-production')
                    } else {
                        echo "No deploy for branch ${env.GIT_BRANCH}"
                    }
                }
            }
        }
    }

    post {
        always {
            echo "Final cleanup"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down || true'
            }
        }

        success {
            echo "Build & deploy pipeline finished: SUCCESS"
        }

        failure {
            echo "Build & deploy pipeline finished: FAILURE"
        }

        aborted {
            echo "Build & deploy pipeline: ABORTED"
        }
    }
}
