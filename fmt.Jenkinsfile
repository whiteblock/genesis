/*
This jenkinsfile will lint and format (fmt) golang source files
*/

pipeline {
  agent {
    kubernetes {
      cloud 'kubernetes-dev-gke'
      yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    cicd: true
spec:
  containers:
  - name: busybox
    image: busybox
    command:
    - cat
    tty: true
"""
    }
  }
  stages {
    stage('Test') {
      steps {
        container('busybox') {
          sh "echo 'hello world!'"
        }
      }
    }
  }
}
