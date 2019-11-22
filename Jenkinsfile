def DEFAULT_BRANCH = 'dev'

pipeline {
  agent any
  environment {
  //General
    APP_NAME              = "genesis"
    GCP_REGION            = credentials('GCP_REGION')
    APP_NAMESPACE         = "apps"
    CHART_PATH            = "kubernetes/helm"
    KUBEDOG_VERSION       = "0.3.2"
    REV_SHORT             = sh(script: "git log --pretty=format:'%h' -n 1", , returnStdout: true).trim()
    INFRA_REPO_URL        = credentials('INFRA_REPO_URL')
    CODECOV_TOKEN         = credentials('CODECOV_TOKEN')
    CI_ENV                = sh(script: "curl -s https://codecov.io/env | bash", , returnStdout: true).trim()

  // Dev
    DEV_GCP_PROJECT_ID    = credentials('DEV_GCP_PROJECT_ID')
    DEV_GKE_CLUSTER_NAME  = credentials('DEV_GKE_CLUSTER_NAME')
    DEV_KUBE_CONTEXT      = "gke_${DEV_GCP_PROJECT_ID}_${GCP_REGION}_${DEV_GKE_CLUSTER_NAME}"
    IMAGE_REPO            = "gcr.io/${DEV_GCP_PROJECT_ID}"

  // Prod
    PROD_GCP_PROJECT_ID   = credentials('PROD_GCP_PROJECT_ID')
    PROD_GKE_CLUSTER_NAME = credentials('PROD_GKE_CLUSTER_NAME')
    PROD_KUBE_CONTEXT     = "gke_${PROD_GCP_PROJECT_ID}_${GCP_REGION}_${PROD_GKE_CLUSTER_NAME}"

  // Slack
    SLACK_CHANNEL         = '#alerts'
    SLACK_CREDENTIALS_ID  = 'jenkins-slack-integration-token'
    SLACK_TEAM_DOMAIN     = 'whiteblock'
    GO111MODULE           = 'on'
  }
  options {
    buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
  }
  stages {
    stage('Run tests') {
      agent {
        docker { 
          image "golang:1.13.4-alpine"
          args  "-u root ${CI_ENV}"
        }
      }
      when {
        beforeAgent true
        anyOf {
          changeRequest target: 'dev'
          changeRequest target: 'master'
        }
      }
      steps {
        script {
          sh "apk add git gcc libc-dev curl bash"
          sh "go get github.com/vektra/mockery/.../"
          sh "go get -u golang.org/x/lint/golint"
          sh "sh tests.sh"
        }
      }
      post {
        success {
          sh ":"
        }
        cleanup {
          sh "chmod -R 777 coverage.txt mocks || true"
          deleteDir()
        }
      }
    }
    stage('Build docker image') {
      when {
        anyOf {
          branch 'master'
          branch 'dev'
        }
      }
      steps {
        script {
          withCredentials([file(credentialsId: 'google-infra-dev-auth', variable: 'GOOGLE_APPLICATION_CREDENTIALS')]) {
            sh """
              gcloud auth activate-service-account --key-file ${GOOGLE_APPLICATION_CREDENTIALS}
              gcloud auth configure-docker
            """
            sh "docker pull ${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME}-build-latest || true"
            docker.build("${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME}-build-latest",
                        "--cache-from ${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME}-build-latest --target build .")
            docker.build("${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME}-${REV_SHORT}",
                        "--cache-from ${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME}-build-latest .")
            docker.image("${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME}-${REV_SHORT}").push()
            docker.image("${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME}-build-latest").push()
          }
        }
      }
    }
    stage('Deploy to dev cluster') {
      when {
        anyOf {
          branch 'dev'
        }
      }
      steps {
        script {
          dir('infra') {
            git branch: 'master',
                credentialsId: 'jenkins-github',
                url: "${INFRA_REPO_URL}"
            withCredentials([file(credentialsId: 'google-infra-dev-auth', variable: 'GOOGLE_APPLICATION_CREDENTIALS')]) {
              sh "gcloud auth activate-service-account --key-file ${GOOGLE_APPLICATION_CREDENTIALS}"
              sh "gcloud config set project ${DEV_GCP_PROJECT_ID}"
              sh "gcloud container clusters get-credentials ${DEV_GKE_CLUSTER_NAME} --project ${DEV_GCP_PROJECT_ID} --region ${GCP_REGION}"
              sh "git secret reveal"
              sh "curl -L https://dl.bintray.com/flant/kubedog/v${KUBEDOG_VERSION}/kubedog-linux-amd64-v${KUBEDOG_VERSION} -o kubedog"
              sh "chmod +x kubedog"
              sh "helm init --kube-context=${DEV_KUBE_CONTEXT} --tiller-namespace helm --service-account tiller --history-max 10 --client-only"
              sh "helm upgrade ${APP_NAME} ${CHART_PATH}/${APP_NAME} --install --namespace ${APP_NAMESPACE} --tiller-namespace helm \
                  --values ${CHART_PATH}/${APP_NAME}/dev.values.yaml \
                  --set-string image.tag=${BRANCH_NAME}-${REV_SHORT} \
                  --set-string image.repository=${IMAGE_REPO}/${APP_NAME}"
              sh "echo '{\"Deployments\": [{\"ResourceName\": \"'${APP_NAME}'\",\"Namespace\": \"'${APP_NAMESPACE}'\"}]}' | ./kubedog multitrack -t 60"
            }
          }
        }
      }
    }
    stage('Deploy to prod cluster') {
      when {
        anyOf {
          branch 'master'
        }
      }
      steps {
        script {
          dir('infra') {
            git branch: 'master',
                credentialsId: 'jenkins-github',
                url: "${INFRA_REPO_URL}"
            withCredentials([file(credentialsId: 'google-infra-dev-auth', variable: 'GOOGLE_APPLICATION_CREDENTIALS')]) {
              sh "gcloud auth activate-service-account --key-file ${GOOGLE_APPLICATION_CREDENTIALS}"
              sh "gcloud config set project ${PROD_GCP_PROJECT_ID}"
              sh "gcloud container clusters get-credentials ${PROD_GKE_CLUSTER_NAME} --project ${PROD_GCP_PROJECT_ID} --region ${GCP_REGION}"
              sh "git secret reveal"
              sh "curl -L https://dl.bintray.com/flant/kubedog/v${KUBEDOG_VERSION}/kubedog-linux-amd64-v${KUBEDOG_VERSION} -o kubedog"
              sh "chmod +x kubedog"
              sh "helm init --kube-context=${PROD_KUBE_CONTEXT} --tiller-namespace helm --service-account tiller --history-max 10 --client-only"
              sh "helm upgrade ${APP_NAME} ${CHART_PATH}/${APP_NAME} --install --namespace ${APP_NAMESPACE} --tiller-namespace helm \
                  --values ${CHART_PATH}/${APP_NAME}/prod.values.yaml \
                  --set-string image.tag=${BRANCH_NAME}-${REV_SHORT} \
                  --set-string image.repository=${IMAGE_REPO}/${APP_NAME}"
              sh "echo '{\"Deployments\": [{\"ResourceName\": \"'${APP_NAME}'\",\"Namespace\": \"'${APP_NAMESPACE}'\"}]}' | ./kubedog multitrack -t 60"
            }
          }
        }
      }
    }
  }
  post {
    failure {
      sh "docker network prune --force"
      sh "docker volume prune --force"
      script {
        if (env.BRANCH_NAME == DEFAULT_BRANCH || env.BRANCH_NAME == 'master') {
          withCredentials([
              [$class: 'StringBinding', credentialsId: "${SLACK_CREDENTIALS_ID}", variable: 'SLACK_TOKEN']
          ]) {
              slackSend teamDomain: "${SLACK_TEAM_DOMAIN}",
                  channel: "${SLACK_CHANNEL}",
                  token: "${SLACK_TOKEN}",
                  color: 'danger',
                  message: "@here ALARM! \n *FAILED*: Job *${env.JOB_NAME}*. \n <${env.RUN_DISPLAY_URL}|*Build Log [${env.BUILD_NUMBER}]*>"
          }
        }
      }
    }
    always {
      sh "kubectl config delete-context ${DEV_KUBE_CONTEXT} || true"
      sh "kubectl config delete-context ${PROD_KUBE_CONTEXT} || true"
      sh "gcloud auth revoke || true"
      deleteDir()
      sh "/usr/bin/docker container prune --force --filter 'until=3h'"
      sh "/usr/bin/docker image prune --force --filter 'until=72h'"
    }
  }
}
