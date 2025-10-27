pipeline {
	agent any
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
				beforeAgent: true
				anyOf {
					branch 'dev'
					expression { return env.TAG_NAME != null }
				}
			}
			steps {
				echo 'Building project...'
				sh 'go mod download'
				sh 'go build -o product cmd/main.go'
				echo 'Build success'
			}
		}
		stage('Test') {
			when {
				beforeAgent: true
				anyOf {
					branch 'dev'
					expression { return env.TAG_NAME != null }
				}
			}
			steps {
				if (env.TAG_NAME) {
					DOCKER_TAG = "${env.TAG_NAME}"
				}
				sh 'if [ ! -d "tests" ]; then
					echo "echo Test skipped"
				else
					echo "go test -v"
				fi'
			}
		}
		stage('Build and Push Docker Image') {
			when {
				beforeAgent: true
				anyOf {
					branch 'dev'
					expression { return env.TAG_NAME != null }
				}
			}
			steps {
				echo 'Building Docker image...'
				sh 'docker build --build-arg CONSUL_HOST=${CONSUL_HOST} --build-arg CONSUL_PORT=${CONSUL_PORT} --build-arg CONSUL_PREFIX=product -t ${DOCKER_IMAGE}:${DOCKER_TAG} .'
				echo 'Build images success'
				echo 'Starting to push image...'
				sh 'docker push ${DOCKER_IMAGE}:${DOCKER_TAG}'
				echo 'Push Image success'
			}
		}
		stage('Deploy to Kubernetes') {
			when {
				beforeAgent: true
				anyOf {
					branch 'dev'
					expression { return env.TAG_NAME != null }
				}
			}
			steps {
				echo 'Starting to deploy...'
				withKubeConfig([credentialsId: 'kubernetes-config', serverUrl: "${KUBERNETES_API_SERVER}", namespace: 'dev']) {
					sh 'kubectl set image deployment/product-service ${DOCKER_IMAGE}:${DOCKER_TAG}'
				}
				echo 'Deploy to kubernetes success'
			}
		}
	}
}