pipeline {
	agent any
	tools {
	    go 'go-1.20.10'
	}
	environment {
		CONSUL_HOST = credentials('CONSUL_HOST')
		CONSUL_PORT = credentials('CONSUL_PORT')
		DOCKER_IMAGE = '192.168.0.62/microservice/product'
		DOCKER_TAG = "${env.GIT_BRANCH}-${env.GIT_COMMIT.substring(0, 8)}"
		KUBERNETES_API_SERVER = credentials('kubernetes-api-server')
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
				sh """
				echo 'Building project...'
				go mod download
				go build -o product cmd/main.go
				echo 'Build success'
				"""
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
					docker.build("${DOCKER_IMAGE}:${DOCKER_TAG}", "--build-arg CONSUL_HOST=${CONSUL_HOST} --build-arg CONSUL_PORT=${CONSUL_PORT} --build-arg CONSUL_PREFIX=product")
					docker.withRegistry('https://192.168.0.62', 'harbor-jenkins') {
						docker.image("${DOCKER_IMAGE}:${DOCKER_TAG}").push()
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
				withKubeConfig([credentialsId: 'kubernetes-config', serverUrl: "${KUBERNETES_API_SERVER}", namespace: 'dev']) {
					sh """
					echo 'Starting to deploy...'
					/usr/bin/kubectl --kubeconfig=$KUBECONFIG set image deployment/product-service ${DOCKER_IMAGE}:${DOCKER_TAG} -n dev
					echo 'Deploy to kubernetes success'
					"""
				}
			}
		}
	}
	post {
		always {
			cleanWs()
		}
		success {
			echo "🎉Pipeline ${DOCKER_IMAGE}:${DOCKER_TAG} deploy succeeded"
		}
		failure {
			echo "❌Pipeline ${DOCKER_IMAGE}:${DOCKER_TAG} deploy failed"
		}
	}
}