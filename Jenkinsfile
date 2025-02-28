pipeline {
    agent any

    environment {
        BE_BINARY_NAME = "be-finance"
        BE_BINARY_PATH_DEV = "/opt/be-binaries/kaffah-priority-finance/dev"
        BE_BINARY_PATH_STAGING = "/opt/be-binaries/kaffah-priority-finance/staging"
        BE_BINARY_PATH_PROD = "/opt/be-binaries/kaffah-priority-finance/prod"
        BE_FINANCE_SERVICE_NAME_DEV = "be-finance-dev.service"
        BE_FINANCE_SERVICE_NAME_STAGING = "be-finance-staging.service"
        BE_FINANCE_SERVICE_NAME_PROD = "be-finance-prod.service"
    }

    stages {
        stage('Build') {
            steps {
                echo "Building for branch " + env.BRANCH_NAME
                sh '/usr/local/go/bin/go mod tidy'
                sh '/usr/local/go/bin/go build -o ${BE_BINARY_NAME} ./cmd/bin/main.go'
            }
        }

        stage('Test') {
            steps {
                echo "Testing for branch " + env.BRANCH_NAME
            }
        }

        stage('Deploy') {
            steps {
                script {
                    if (env.BRANCH_NAME == 'dev') {
                        echo "Deploying to DEV"
                        sh """
                        mkdir -p ${BE_BINARY_PATH_DEV}
                        sudo systemctl stop ${BE_FINANCE_SERVICE_NAME_DEV}
                        cp ${BE_BINARY_NAME} ${BE_BINARY_PATH_DEV}
                        sudo systemctl start ${BE_FINANCE_SERVICE_NAME_DEV}
                        """
                    } else if (env.BRANCH_NAME == 'staging') {
                        echo "Deploying to STAGING"
                        sh """
                        mkdir -p ${BE_BINARY_PATH_STAGING}
                        sudo systemctl stop ${BE_FINANCE_SERVICE_NAME_STAGING}
                        cp ${BE_BINARY_NAME} ${BE_BINARY_PATH_STAGING}
                        sudo systemctl start ${BE_FINANCE_SERVICE_NAME_STAGING}
                        """
                    } else if (env.BRANCH_NAME == 'production') {
                        echo "Deploying to PRODUCTION"
                        sh """
                        mkdir -p ${BE_BINARY_PATH_PROD}
                        sudo systemctl stop ${BE_FINANCE_SERVICE_NAME_PROD}
                        cp ${BE_BINARY_NAME} ${BE_BINARY_PATH_PROD}
                        sudo systemctl start ${BE_FINANCE_SERVICE_NAME_PROD}
                        """
                    } else {
                        echo "No deployment for branch ${env.ACTUAL_BRANCH}"
                    }
                }
            }
        }
    }
}
