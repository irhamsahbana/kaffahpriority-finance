pipeline {
    agent any

    environment {
        BE_BINARY_NAME = "be-finance"
        BE_BINARY_PATH_DEV = "/opt/be-binaries/kaffah-priority-finance/dev"
        BE_BINARY_PATH_PROD = "/opt/be-binaries/kaffah-priority-finance/prod"
        BE_FINANCE_SERVICE_NAME_DEV = "be-finance-dev.service"
        BE_FINANCE_SERVICE_NAME_PROD = "be-finance-prod.service"
    }

    stages {
        stage('Initialize') {
            steps {
                echo "Pipeline started"
            }
        }

        stage('Determine Branch') {
            steps {
                script {
                    // Pastikan kita berada di branch yang benar
                    sh "git fetch --all"
                    sh "git checkout $(git rev-parse --abbrev-ref HEAD) || true"

                    // Ambil nama branch dengan cara yang lebih aman
                    env.ACTUAL_BRANCH = sh(script: "git symbolic-ref --short HEAD || git rev-parse --abbrev-ref HEAD", returnStdout: true).trim()

                    echo "Detected branch: ${env.ACTUAL_BRANCH}"
                }
            }
        }

        stage('Build') {
            steps {
                echo "Building for branch ${env.ACTUAL_BRANCH}"
                sh '/usr/local/go/bin/go mod tidy'
                sh '/usr/local/go/bin/go build -o ${BE_BINARY_NAME} ./cmd/bin/main.go'
            }
        }

        stage('Test') {
            steps {
                echo "Testing for branch ${env.ACTUAL_BRANCH}"
                // Tambahkan perintah test jika diperlukan
            }
        }

        stage('Deploy') {
            steps {
                script {
                    if (env.ACTUAL_BRANCH == 'dev') {
                        echo "Deploying to DEV"
                        sh """
                        mkdir -p ${BE_BINARY_PATH_DEV}
                        cp ${BE_BINARY_NAME} ${BE_BINARY_PATH_DEV}
                        systemctl stop ${BE_FINANCE_SERVICE_NAME_DEV}
                        systemctl start ${BE_FINANCE_SERVICE_NAME_DEV}
                        """
                    } else if (env.ACTUAL_BRANCH == 'production') {
                        echo "Deploying to PRODUCTION"
                        sh """
                        mkdir -p ${BE_BINARY_PATH_PROD}
                        cp ${BE_BINARY_NAME} ${BE_BINARY_PATH_PROD}
                        systemctl stop ${BE_FINANCE_SERVICE_NAME_PROD}
                        systemctl start ${BE_FINANCE_SERVICE_NAME_PROD}
                        """
                    } else {
                        echo "No deployment for branch ${env.ACTUAL_BRANCH}"
                    }
                }
            }
        }
    }
}
