@Library('whiteblock-dev')_

def DEFAULT_BRANCH = 'master'

pipeline {
  agent any
  environment {
  //General
    APP_NAME              = "genesis"
    GCP_REGION            = credentials('GCP_REGION')
    APP_NAMESPACE         = "apps"
    CHART_PATH            = "chart"
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
  }
  options {
    disableConcurrentBuilds()
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
          changeRequest target: DEFAULT_BRANCH
        }
      }
      steps {
        script {
          sh "apk add git gcc libc-dev curl bash make"
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
      when { branch DEFAULT_BRANCH }
      steps {
        script {
          withCredentials([file(credentialsId: 'google-infra-dev-auth', variable: 'GOOGLE_APPLICATION_CREDENTIALS')]) {
            sh """
              gcloud auth activate-service-account --key-file ${GOOGLE_APPLICATION_CREDENTIALS}
              gcloud auth configure-docker
            """
            sh "docker pull ${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME} || true"
            docker.build("${IMAGE_REPO}/${APP_NAME}",
                        "--cache-from ${IMAGE_REPO}/${APP_NAME}:${BRANCH_NAME} .")
            docker.image("${IMAGE_REPO}/${APP_NAME}").push("${BRANCH_NAME}")
            docker.image("${IMAGE_REPO}/${APP_NAME}").push("${REV_SHORT}")
          }
        }
      }
    }
    stage('Deploy to dev cluster') {
      when { branch DEFAULT_BRANCH }
      steps {
        script {
          withCredentials([file(credentialsId: 'google-infra-dev-auth', variable: 'GOOGLE_APPLICATION_CREDENTIALS')]) {
            sh "gcloud auth activate-service-account --key-file ${GOOGLE_APPLICATION_CREDENTIALS}"
            sh "gcloud config set project ${DEV_GCP_PROJECT_ID}"
            sh "gcloud container clusters get-credentials ${DEV_GKE_CLUSTER_NAME} --project ${DEV_GCP_PROJECT_ID} --region ${GCP_REGION}"
            sh "curl -L https://dl.bintray.com/flant/kubedog/v${KUBEDOG_VERSION}/kubedog-linux-amd64-v${KUBEDOG_VERSION} -o kubedog"
            sh "chmod +x kubedog"
            sh "helm init --kube-context=${DEV_KUBE_CONTEXT} --tiller-namespace helm --service-account tiller --history-max 10 --client-only"
            sh "helm upgrade ${APP_NAME} ${CHART_PATH}/${APP_NAME} --install --namespace ${APP_NAMESPACE} --tiller-namespace helm \
                --set-string image.tag=${REV_SHORT} \
                --set-string image.repository=${IMAGE_REPO}/${APP_NAME}"
            sh "echo '{\"Deployments\": [{\"ResourceName\": \"'${APP_NAME}'\",\"Namespace\": \"'${APP_NAMESPACE}'\"}]}' | ./kubedog multitrack -t 60"
          }
        }
      }
    }
  }
  post {
    failure {
      sh "docker network prune --force"
      sh "docker volume prune --force"
    }
    always {
      // slack notify first to ignore teardown errors
      slackNotify(env.BRANCH_NAME == DEFAULT_BRANCH)
      sh "kubectl config delete-context ${DEV_KUBE_CONTEXT} || true"
      sh "gcloud auth revoke || true"
      deleteDir()
      sh "/usr/bin/docker container prune --force --filter 'until=3h'"
      sh "/usr/bin/docker image prune --force --filter 'until=72h'"
    }
  }
}
