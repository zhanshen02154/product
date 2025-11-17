pipeline {
	agent any
	tools {
	    go 'go-1.20.10'
	}
	environment {
		DOCKER_IMAGE = '192.168.0.62/microservice/product'
		DOCKER_TAG = "${env.GIT_BRANCH}-${env.GIT_COMMIT.substring(0, 8)}"
		GOPROXY = 'https://goproxy.cn,direct'
	}
	stages {
		stage('Build') {
			when {
				anyOf {
					branch 'dev'
					expression { return env.TAG_NAME != null }
				}
			}
			steps {
				sh '''
				echo 'Building project...'
				export CGO_ENABLED=0
				export GOOS=linux
				export GOARCH=amd64
				go env -w GO111MODULE=on
				go mod download
				go build -o product cmd/main.go
				echo 'Build success'
				'''
			}
		}
		stage('Build and Push Docker Image') {
			when {
				anyOf {
					branch 'dev'
					expression { return env.TAG_NAME != null }
				}
			}
			steps {
				script {
					if (env.TAG_NAME) {
						DOCKER_TAG = "${env.TAG_NAME}"
					}
					withCredentials([string(credentialsId: 'CONSUL_HOST', variable: 'consul_host'), 
						string(credentialsId: 'CONSUL_PORT', variable: 'consul_port')
						]) {
						sh 'set +x'
						docker.build("${DOCKER_IMAGE}:${DOCKER_TAG}", "--build-arg CONSUL_HOST=$consul_host --build-arg CONSUL_PORT=$consul_port --build-arg CONSUL_PREFIX=product .")
						docker.withRegistry('https://192.168.0.62', 'harbor-jenkins') {
							docker.image("${DOCKER_IMAGE}:${DOCKER_TAG}").push()
						}
						sh '''
						docker rmi ${DOCKER_IMAGE}:${DOCKER_TAG}
						'''
					}
				}
			}
		}
		stage('Deploy to Kubernetes') {
			when {
				anyOf {
					branch 'dev'
					expression { return env.TAG_NAME != null }
				}
			}
			steps {
				script {
					try {
						withCredentials([string(credentialsId: 'kubernetes-api-server', variable: 'k8s_api_server')]) {
							withKubeConfig([credentialsId: 'kubernetes-config', serverUrl: "$k8s_api_server", namespace: 'dev']) {
								sh '''
								set +x
								/usr/bin/kubectl set image deployment/product-service product-container=${DOCKER_IMAGE}:${DOCKER_TAG} -n dev --record
								/usr/bin/kubectl rollout status deployment/product-service -n dev --timeout 120s
								'''
							}
						}
					}
					catch(exc) {
						withCredentials([string(credentialsId: 'kubernetes-api-server', variable: 'k8s_api_server')]) {
							withKubeConfig([credentialsId: 'kubernetes-config', serverUrl: "$k8s_api_server", namespace: 'dev']) {
								sh '''
								set +x
								/usr/bin/kubectl rollout undo deployment/product-service -n dev --to-revision=$(/usr/bin/kubectl rollout history deployment/product-service -n dev | awk '{print $1}' | tail -n 2)
								'''
							}
						}
					}
				}
			}
		}
	}
	post {
		always {
			deleteDir()
		}
		success {
			echo "🎉Pipeline ${DOCKER_IMAGE}:${DOCKER_TAG} deploy succeeded"
		}
		failure {
			echo "❌Pipeline ${DOCKER_IMAGE}:${DOCKER_TAG} deploy failed"
		}
	}
}