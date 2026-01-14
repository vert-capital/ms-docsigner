@NonCPS
def cancelPreviousBuilds() {
  String jobname = env.JOB_NAME
  int currentBuildNum = env.BUILD_NUMBER.toInteger()

  def job = Jenkins.instance.getItemByFullName(jobname)
  for (build in job.builds) {
    if (build.isBuilding() && currentBuildNum > build.getNumber().toInteger()) {
      build.doStop();
      echo "Build ${build.toString()} cancelled"
    }
  }
}

def bitbucketNotify(status, branch_name, git_commit) {
    withCredentials([usernamePassword(credentialsId: 'thiagofreitas', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
        sh "curl --location --request POST 'https://api.bitbucket.org/2.0/repositories/sistema_vert/vertc-gerador-documentos-backend/commit/"+git_commit+"/statuses/build'" \
            + " --user $USERNAME:$PASSWORD " \
            + " --header 'Content-Type: application/json' " \
            + " --data '{" \
            + "    \"state\": \""+status+"\"," \
            + "    \"key\": \""+git_commit+"\"," \
            + "    \"name\": \"Jenkins: "+branch_name+"\"," \
            + "    \"url\": \"https://ci.vert-capital.com/blue/organizations/jenkins/ms-docsigner/activity\"" \
            + "}'"
    }
}


def buildAndPushDockerImage(String environment) {
    def imageNameMap = [
        'develop': 'ms-docsigner-stg',
        'homolog': 'ms-docsigner-hml',
        'master': 'ms-docsigner-prd'
    ]

    def imageName = imageNameMap[environment]

    docker.withRegistry("https://$registry", registryCredential) {

        dockerImage = docker.build(imageName, "--cache-from $registry/$imageName:latest ./src")
        dockerImage.push("$BUILD_NUMBER")
        dockerImage.push("latest")
    }

    sh "docker rmi $registry/$imageName:$BUILD_NUMBER || true"
    sh "docker rmi $registry/$imageName:latest || true"
}


def deployToEnvironment(String environment) {
    def appNameMap = [
        'develop': 'ms-docsigner-stg',
        'homolog': 'ms-docsigner-hml',
        'master': 'ms-docsigner-prd'
    ]

    def credentialIdMap = [
        'develop': 'argocd-homolog',
        'homolog': 'argocd-homolog',
        'master': 'argocd-production'
    ]

    def appName = appNameMap[environment]
    def credentialId = credentialIdMap[environment]

    withCredentials([string(credentialsId: "ARGOCD_SERVER", variable: 'ARGOCD_SERVER')]) {
        withCredentials([string(credentialsId: credentialId, variable: 'ARGOCD_AUTH_TOKEN')]) {
            sh "argocd --grpc-web app actions run $appName restart --kind Deployment --all"
        }
    }
}

pipeline {
    environment {
        registry = "197272534240.dkr.ecr.us-east-1.amazonaws.com"
        registryCredential = "ecr:us-east-1:aws_vert"
        dockerImageName = ""

        AWS_ACCESS_KEY_ID = credentials("GERDOC_BACK_ACCESS_KEY")
        AWS_SECRET_ACCESS_KEY = credentials("GERDOC_BACK_SECRET_ACCESS")
        AWS_STORAGE_BUCKET_NAME = credentials("GERDOC_BACK_BUCKET_NAME")
        AWS_LOCATION = credentials("GERDOC_BACK_REGION")
        INGESTAO_S3_BUCKET_NAME = credentials("GERDOC_BACK_INGEST_BUCKET_NAME")
        INGESTAO_S3_REGION_NAME = credentials("GERDOC_BACK_INGEST_REGION")
        INGESTAO_S3_ACCESS_KEY_ID = credentials("GERDOC_BACK_INGEST_ACCESS_KEY")
        INGESTAO_S3_SECRET_ACCESS_KEY = credentials("GERDOC_BACK_INGEST_SECRET_ACCESS")

        MS_INGEST_URL = "http://host.docker.internal:8084/"
        MS_INGEST_RETRY_COUNT = "3"
        MS_INGEST_RETRY_DELAY = "5"
        MS_INGEST_MAX_BACKOFF = "60"
        KAFKA_DLQ_TOPIC = "gerador-documentos-dlq"
        NOTIFICATION_EMAIL = "notificacoes@teste.com"
        FALLBACK_SAMPLE_ENABLED = "True"
        OTEL_ENABLED = "False"
        MS_DOCSIGNER_URL = "http://host.docker.internal:8080"
        MS_DOCSIGNER_TIMEOUT = "180"
        MS_DOCSIGNER_RETRY_COUNT = "3"
        MS_DOCSIGNER_RETRY_DELAY = "10"
        MS_DOCSIGNER_MAX_BACKOFF = "60"
        MS_DOCSIGNER_USER_EMAIL = "root@root.com.br"
        MS_DOCSIGNER_USER_PASSWORD = "root"
        MS_DOCSIGNER_AUTO_SIGNATURE_ADMIN_EMAIL = "root@gmail.com"
        MS_DOCSIGNER_AUTO_SIGNATURE_API_EMAIL = "root@gmail.com"
        POSTGRES_DB = "gerador_documentos"
        POSTGRES_USER = "root"
        POSTGRES_PASSWORD = "root"
        POSTGRES_HOST = "db"
        POSTGRES_PORT = "5432"
        DEBUG = "True"
        LOCAL_ENV = "False"
        EMAIL_HOST = "mail"
        EMAIL_HOST_USER = ""
        EMAIL_HOST_PASSWORD = ""
        EMAIL_PORT = "1025"
        EMAIL_USE_TLS = "False"
        EMAIL_FROM = "teste@teste.com"
        KAFKA_BOOTSTRAP_SERVER = "kafka:29092"
        KAFKA_CLIENT_ID = "gerador-documentos-back"
        KAFKA_GROUP_ID = "gerador-documentos-back"
        CAS_SERVER_URL = "https://sso.stg.vert-tech.dev/cas/"
        FRONTEND_AUTH_REDIRECT = ""
        WEBHOOK_SECRET_TOKEN = ""
        REDIS_HOST = "redis"
        REDIS_DB = "1"
        REDIS_PORT = "6379"
        MSBOT_KEY = ""
        MSBOT_URL = ""
        MSBOT_ACTIVE = "False"
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
                    if (!env.BRANCH_NAME) {
                        env.BRANCH_NAME = rawBranch
                    }


                    bitbucketNotify('INPROGRESS', env.BRANCH_NAME, env.GIT_COMMIT)
                }
            }
        }

        stage('Code Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Docker Images') {
            steps {
                script {

                    sh """
                    rm -f src/.env
                    touch src/.env
                    echo AWS_ACCESS_KEY_ID='${env.AWS_ACCESS_KEY_ID}' >> src/.env
                    echo AWS_SECRET_ACCESS_KEY='${env.AWS_SECRET_ACCESS_KEY}' >> src/.env
                    echo AWS_STORAGE_BUCKET_NAME='${env.AWS_STORAGE_BUCKET_NAME}' >> src/.env
                    echo AWS_LOCATION='${env.AWS_LOCATION}' >> src/.env
                    echo INGESTAO_S3_BUCKET_NAME='${env.INGESTAO_S3_BUCKET_NAME}' >> src/.env
                    echo INGESTAO_S3_REGION_NAME='${env.INGESTAO_S3_REGION_NAME}' >> src/.env
                    echo INGESTAO_S3_ACCESS_KEY_ID='${env.INGESTAO_S3_ACCESS_KEY_ID}' >> src/.env
                    echo INGESTAO_S3_SECRET_ACCESS_KEY='${env.INGESTAO_S3_SECRET_ACCESS_KEY}' >> src/.env
                    echo MS_INGEST_URL='${env.MS_INGEST_URL}' >> src/.env
                    echo MS_INGEST_RETRY_COUNT='${env.MS_INGEST_RETRY_COUNT}' >> src/.env
                    echo MS_INGEST_RETRY_DELAY='${env.MS_INGEST_RETRY_DELAY}' >> src/.env
                    echo MS_INGEST_MAX_BACKOFF='${env.MS_INGEST_MAX_BACKOFF}' >> src/.env
                    echo KAFKA_DLQ_TOPIC='${env.KAFKA_DLQ_TOPIC}' >> src/.env
                    echo NOTIFICATION_EMAIL='${env.NOTIFICATION_EMAIL}' >> src/.env
                    echo FALLBACK_SAMPLE_ENABLED='${env.FALLBACK_SAMPLE_ENABLED}' >> src/.env
                    echo OTEL_ENABLED='${env.OTEL_ENABLED}' >> src/.env
                    echo MS_DOCSIGNER_URL='${env.MS_DOCSIGNER_URL}' >> src/.env
                    echo MS_DOCSIGNER_TIMEOUT='${env.MS_DOCSIGNER_TIMEOUT}' >> src/.env
                    echo MS_DOCSIGNER_RETRY_COUNT='${env.MS_DOCSIGNER_RETRY_COUNT}' >> src/.env
                    echo MS_DOCSIGNER_RETRY_DELAY='${env.MS_DOCSIGNER_RETRY_DELAY}' >> src/.env
                    echo MS_DOCSIGNER_MAX_BACKOFF='${env.MS_DOCSIGNER_MAX_BACKOFF}' >> src/.env
                    echo MS_DOCSIGNER_USER_EMAIL='${env.MS_DOCSIGNER_USER_EMAIL}' >> src/.env
                    echo MS_DOCSIGNER_USER_PASSWORD='${env.MS_DOCSIGNER_USER_PASSWORD}' >> src/.env
                    echo MS_DOCSIGNER_AUTO_SIGNATURE_ADMIN_EMAIL='${env.MS_DOCSIGNER_AUTO_SIGNATURE_ADMIN_EMAIL}' >> src/.env
                    echo MS_DOCSIGNER_AUTO_SIGNATURE_API_EMAIL='${env.MS_DOCSIGNER_AUTO_SIGNATURE_API_EMAIL}' >> src/.env
                    echo POSTGRES_DB='${env.POSTGRES_DB}' >> src/.env
                    echo POSTGRES_USER='${env.POSTGRES_USER}' >> src/.env
                    echo POSTGRES_PASSWORD='${env.POSTGRES_PASSWORD}' >> src/.env
                    echo POSTGRES_HOST='${env.POSTGRES_HOST}' >> src/.env
                    echo POSTGRES_PORT='${env.POSTGRES_PORT}' >> src/.env
                    echo DEBUG='${env.DEBUG}' >> src/.env
                    echo LOCAL_ENV='False' >> src/.env
                    echo EMAIL_HOST='${env.EMAIL_HOST}' >> src/.env
                    echo EMAIL_HOST_USER='${env.EMAIL_HOST_USER}' >> src/.env
                    echo EMAIL_HOST_PASSWORD='${env.EMAIL_HOST_PASSWORD}' >> src/.env
                    echo EMAIL_PORT='${env.EMAIL_PORT}' >> src/.env
                    echo EMAIL_USE_TLS='${env.EMAIL_USE_TLS}' >> src/.env
                    echo EMAIL_FROM='${env.EMAIL_FROM}' >> src/.env
                    echo KAFKA_BOOTSTRAP_SERVER='${env.KAFKA_BOOTSTRAP_SERVER}' >> src/.env
                    echo KAFKA_CLIENT_ID='${env.KAFKA_CLIENT_ID}' >> src/.env
                    echo KAFKA_GROUP_ID='${env.KAFKA_GROUP_ID}' >> src/.env
                    echo CAS_SERVER_URL='${env.CAS_SERVER_URL}' >> src/.env
                    echo FRONTEND_AUTH_REDIRECT='${env.FRONTEND_AUTH_REDIRECT}' >> src/.env
                    echo WEBHOOK_SECRET_TOKEN='${env.WEBHOOK_SECRET_TOKEN}' >> src/.env
                    echo REDIS_HOST='${env.REDIS_HOST}' >> src/.env
                    echo REDIS_DB='${env.REDIS_DB}' >> src/.env
                    echo REDIS_PORT='${env.REDIS_PORT}' >> src/.env
                    echo MSBOT_KEY='${env.MSBOT_KEY}' >> src/.env
                    echo MSBOT_URL='${env.MSBOT_URL}' >> src/.env
                    echo MSBOT_ACTIVE='${env.MSBOT_ACTIVE}' >> src/.env
                    """

                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml build'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml up -d --no-build'
                }
            }
        }
        
        stage('Parallel Tests') {
            failFast true
            parallel {
                 

                stage('SonarQube analysis') {
                    when {
                        expression {
                            return env.GIT_BRANCH == 'master' || env.GIT_BRANCH == 'develop' || env.GIT_BRANCH == 'homolog'
                        }
                    }

                    environment {
                        scannerHome = tool 'SonarQubeScanner'
                    }
                    steps {
                        withSonarQubeEnv(installationName: 'vert-sonar') {
                            sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME} -X"
                        }
                    }
                }
            }
        }

        stage('stop containers') {
            steps {
                script {
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                }
            }
        }

        stage('Build and Deploy') {
            when {
                expression {
                    return env.GIT_BRANCH == 'master' || env.GIT_BRANCH == 'develop' || env.GIT_BRANCH == 'homolog'
                }
            }
            steps {
                script {

                    buildAndPushDockerImage(env.GIT_BRANCH)


                    deployToEnvironment(env.GIT_BRANCH)
                }
            }
        }

    }

    post {
        always {
            echo "Stop Docker image"
            script{
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
            }
        }

        success {
            echo "Notify bitbucket success"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                bitbucketNotify('SUCCESSFUL', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }

        failure {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                bitbucketNotify('FAILED', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }

        aborted {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                bitbucketNotify('FAILED', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }

        unsuccessful {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                bitbucketNotify('FAILED', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }
    }
}
