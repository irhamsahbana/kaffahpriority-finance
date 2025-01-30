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
                    sh 'git fetch --all'

                    // Pastikan kita tidak dalam detached HEAD
                    sh '''
                        BRANCH_NAME=$(git symbolic-ref --short HEAD 2>/dev/null || git rev-parse --abbrev-ref HEAD)
                        if [ "$BRANCH_NAME" = "HEAD" ]; then
                            echo "Currently in detached HEAD. Checking out to the correct branch..."
                            BRANCH_NAME=$(git for-each-ref --format="%(refname:short)" refs/remotes/origin/ | grep -E "^(dev|production)$" | head -n 1)
                            if [ -n "$BRANCH_NAME" ]; then
                                git checkout -B $BRANCH_NAME origin/$BRANCH_NAME
                            else
                                echo "Failed to determine the correct branch."
                                exit 1
                            fi
                        fi
                        echo "Final detected branch: $BRANCH_NAME"
                        echo "ACTUAL_BRANCH=$BRANCH_NAME" >> $WORKSPACE/branch.env
                    '''

                    // Load branch name from file
                    env.ACTUAL_BRANCH = sh(script: 'cat $WORKSPACE/branch.env', returnStdout: true).trim()
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
