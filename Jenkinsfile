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
        stage('Build') {
            steps {
                echo "Building for branch ${env.BRANCH_NAME}"
                sh '/usr/local/go/bin/go mod tidy'
                sh '/usr/local/go/bin/go build -o ${BE_BINARY_NAME} ./cmd/bin/main.go'
            }
        }
        stage('Test') {
            steps {
                echo "Testing for branch ${env.BRANCH_NAME}"
                // Tambahkan test script jika ada
            }
        }
        stage('Deploy') {
            steps {
                script {
                    if (env.BRANCH_NAME == 'dev') {
                        echo "Deploying to DEV"
                        sh """
                        mkdir -p ${BE_BINARY_PATH_DEV}
                        cp ${BE_BINARY_NAME} ${BE_BINARY_PATH_DEV}
                        systemctl stop ${BE_FINANCE_SERVICE_NAME_DEV}
                        systemctl start ${BE_FINANCE_SERVICE_NAME_DEV}
                        """
                    } else if (env.BRANCH_NAME == 'production') {
                        echo "Deploying to PRODUCTION"
                        sh """
                        mkdir -p ${BE_BINARY_PATH_PROD}
                        cp ${BE_BINARY_NAME} ${BE_BINARY_PATH_PROD}
                        systemctl stop ${BE_FINANCE_SERVICE_NAME_PROD}
                        systemctl start ${BE_FINANCE_SERVICE_NAME_PROD}
                        """
                    } else {
                        echo "No deployment for branch ${env.BRANCH_NAME}"
                    }
                }
            }
        }
    }
}
