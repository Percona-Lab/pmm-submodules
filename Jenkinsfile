library changelog: false, identifier: 'lib@master', retriever: modernSCM([
    $class: 'GitSCMSource',
    remote: 'https://github.com/Percona-Lab/jenkins-pipelines.git'
]) _

void runAPItests(String DOCKER_IMAGE_VERSION, BRANCH_NAME, GIT_COMMIT_HASH, CLIENT_VERSION, OWNER) {
    stagingJob = build job: 'pmm2-api-tests', parameters: [
        string(name: 'DOCKER_VERSION', value: DOCKER_IMAGE_VERSION),
        string(name: 'GIT_BRANCH', value: BRANCH_NAME),
        string(name: 'OWNER', value: OWNER),
        string(name: 'GIT_COMMIT_HASH', value: GIT_COMMIT_HASH)
    ]
}

void runTestSuite(String DOCKER_IMAGE_VERSION, CLIENT_VERSION, PMM_QA_GIT_BRANCH, PMM_QA_GIT_COMMIT_HASH) {
    stagingJob = build job: 'pmm2-testsuite', parameters: [
        string(name: 'DOCKER_VERSION', value: DOCKER_IMAGE_VERSION),
        string(name: 'CLIENT_VERSION', value: CLIENT_VERSION),
        string(name: 'PMM_QA_GIT_BRANCH', value: PMM_QA_GIT_BRANCH),
        string(name: 'PMM_QA_GIT_COMMIT_HASH', value: PMM_QA_GIT_COMMIT_HASH)
    ]
}

void runUItests(String DOCKER_IMAGE_VERSION, CLIENT_VERSION, PMM_QA_GIT_BRANCH, PMM_QA_GIT_COMMIT_HASH) {
    stagingJob = build job: 'pmm2-ui-tests', parameters: [
        string(name: 'DOCKER_VERSION', value: DOCKER_IMAGE_VERSION),
        string(name: 'CLIENT_VERSION', value: CLIENT_VERSION),
        string(name: 'GIT_BRANCH', value: PMM_QA_GIT_BRANCH),
        string(name: 'GIT_COMMIT_HASH', value: PMM_QA_GIT_COMMIT_HASH)
    ]
}

def isBranchBuild = true
if ( env.CHANGE_URL ) {
    isBranchBuild = false
}

pipeline {
    agent {
        label 'large-amazon'
    }
    stages {
        stage('Prepare') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                sh '''
                    curdir=$(pwd)
                    cd ../
                    wget https://github.com/git-lfs/git-lfs/releases/download/v2.7.1/git-lfs-linux-amd64-v2.7.1.tar.gz
                    tar -zxvf git-lfs-linux-amd64-v2.7.1.tar.gz
                    sudo ./install.sh
                    cd $curdir
                    sudo rm -rf results tmp || :
                    git reset --hard
                    git clean -fdx
                    git submodule foreach --recursive git reset --hard
                    git submodule foreach --recursive git clean -fdx
                    git submodule status
                    export commit_sha=$(git submodule status | grep 'pmm-managed' | awk -F ' ' '{print $1}')
                    curl -s https://api.github.com/repos/percona/pmm-managed/commits/${commit_sha} | grep 'name' | awk -F '"' '{print $4}' | head -1 > OWNER
                    cd sources/pmm-server-packaging/
                    git lfs install
                    git lfs pull
                    git lfs checkout
                    cd $curdir
                    export api_tests_commit_sha=$(git submodule status | grep 'pmm-api-tests' | awk -F ' ' '{print $1}')
                    export api_tests_branch=$(git config -f .gitmodules submodule.pmm-api-tests.branch)
                    echo $api_tests_commit_sha > apiCommitSha
                    echo $api_tests_branch > apiBranch
                    cat apiBranch
                    export pmm_qa_commit_sha=$(git submodule status | grep 'pmm-qa' | awk -F ' ' '{print $1}')
                    export pmm_qa_branch=$(git config -f .gitmodules submodule.pmm-qa.branch)
                    echo $pmm_qa_branch > pmmQABranch
                    echo $pmm_qa_commit_sha > pmmQACommitSha
                    export pmm_ui_tests_commit_sha=$(git submodule status | grep 'grafana-dashboards' | awk -F ' ' '{print $1}')
                    export pmm_ui_tests_branch=$(git config -f .gitmodules submodule.grafana-dashboards.branch)
                    echo $pmm_ui_tests_branch > pmmUITestBranch
                    echo $pmm_ui_tests_commit_sha > pmmUITestsCommitSha
                    cd $curdir
                '''
                installDocker()
                stash includes: 'apiBranch', name: 'apiBranch'
                stash includes: 'pmmQABranch', name: 'pmmQABranch'
                stash includes: 'apiCommitSha', name: 'apiCommitSha'
                stash includes: 'pmmQACommitSha', name: 'pmmQACommitSha'
                stash includes: 'pmmUITestBranch', name: 'pmmUITestBranch'
                stash includes: 'pmmUITestsCommitSha', name: 'pmmUITestsCommitSha'
                slackSend channel: '#pmm-ci', color: '#FFFF00', message: "[${JOB_NAME}]: build started - ${BUILD_URL}"
            }
        }
        stage('Build client source') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                sh '''
                    sg docker -c "
                        env
                        ./build/bin/build-client-source
                    "
                '''
            }
        }
        stage('Build client binary') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                sh '''
                    sg docker -c "
                        env
                        ./build/bin/build-client-binary
                    "
                '''
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
                    sh '''
                        aws s3 cp \
                            --acl public-read \
                            results/tarball/pmm2-client-*.tar.gz \
                            s3://pmm-build-cache/PR-BUILDS/pmm2-client/pmm2-client-${BRANCH_NAME}-${GIT_COMMIT:0:7}.tar.gz
                    '''
                }
                script {
                    def clientPackageURL = sh script:'echo "https://s3.us-east-2.amazonaws.com/pmm-build-cache/PR-BUILDS/pmm2-client/pmm2-client-${BRANCH_NAME}-${GIT_COMMIT:0:7}.tar.gz" | tee CLIENT_URL', returnStdout: true
                    env.CLIENT_URL = sh(returnStdout: true, script: "cat CLIENT_URL").trim()
                }
                stash includes: 'CLIENT_URL', name: 'CLIENT_URL'
            }
        }
        stage('Build client source rpm') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                sh 'sg docker -c "./build/bin/build-client-srpm centos:6"'
            }
        }
        stage('Build client binary rpm') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                sh '''
                    sg docker -c "
                        ./build/bin/build-client-rpm centos:7

                        mkdir -p tmp/pmm-server/RPMS/
                        cp results/rpm/pmm2-client-*.rpm tmp/pmm-server/RPMS/
                    "
                '''
            }
        }
        stage('Build client docker') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'hub.docker.com', passwordVariable: 'PASS', usernameVariable: 'USER')]) {
                    sh """
                        sg docker -c "
                            docker login -u "${USER}" -p "${PASS}"
                        "
                    """
                }
                sh '''
                    sg docker -c "
                        export PUSH_DOCKER=1
                        export DOCKER_CLIENT_TAG=perconalab/pmm-client-fb:${BRANCH_NAME}-${GIT_COMMIT:0:7}

                        ./build/bin/build-client-docker
                    "
                '''
                stash includes: 'results/docker/CLIENT_TAG', name: 'CLIENT_IMAGE'
                archiveArtifacts 'results/docker/CLIENT_TAG'
            }
        }
        stage('Build server packages') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
                    sh '''
                        sg docker -c "
                            export RPM_EPOCH=1
                            export PATH=$PATH:$(pwd -P)/build/bin

                            # 1st-party
                            build-server-rpm percona-dashboards grafana-dashboards
                            build-server-rpm pmm-managed
                            build-server-rpm percona-qan-api2 qan-api2
                            build-server-rpm percona-qan-app qan-app
                            build-server-rpm pmm-server
                            build-server-rpm pmm-update

                            # 3rd-party
                            build-server-rpm clickhouse
                            build-server-rpm prometheus
                            build-server-rpm alertmanager
                            build-server-rpm grafana
                        "
                    '''
                }
            }
        }
        stage('Build server docker') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'hub.docker.com', passwordVariable: 'PASS', usernameVariable: 'USER')]) {
                    sh """
                        sg docker -c "
                            docker login -u "${USER}" -p "${PASS}"
                        "
                    """
                }
                sh '''
                    sg docker -c "
                        export PUSH_DOCKER=1
                        export DOCKER_TAG=perconalab/pmm-server-fb:${BRANCH_NAME}-${GIT_COMMIT:0:7}

                        ./build/bin/build-server-docker
                    "
                '''
                stash includes: 'results/docker/TAG', name: 'IMAGE'
                archiveArtifacts 'results/docker/TAG'
            }
        }
        stage('Tests Execution') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            parallel {
                stage ('Generate FB tags'){
                    steps{
                        script{
                            withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
                                unstash 'IMAGE'
                                def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                                def CLIENT_IMAGE = sh(returnStdout: true, script: "cat results/docker/CLIENT_TAG").trim()
                                sh """
                                    set -o xtrace
                                    curl -v -X POST \
                                        -H "Authorization: token ${GITHUB_API_TOKEN}" \
                                        -d "{\\"body\\":\\"server docker - ${IMAGE}\\nclient docker - ${CLIENT_IMAGE}\\nclient - https://s3.us-east-2.amazonaws.com/pmm-build-cache/PR-BUILDS/pmm2-client/pmm2-client-${BRANCH_NAME}-\${GIT_COMMIT:0:7}.tar.gz\\"}" \
                                        "https://api.github.com/repos/\$(echo $CHANGE_URL | cut -d '/' -f 4-5)/issues/${CHANGE_ID}/comments"
                                """
                            }
                        }
                    }
                }
                stage('Test: API') {
                    steps {
                        script {
                            unstash 'IMAGE'
                            unstash 'apiBranch'
                            unstash 'apiCommitSha'
                            def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                            def CLIENT_IMAGE = sh(returnStdout: true, script: "cat results/docker/CLIENT_TAG").trim()
                            def OWNER = sh(returnStdout: true, script: "cat OWNER").trim()
                            def CLIENT_URL = sh(returnStdout: true, script: "cat CLIENT_URL").trim()
                            def API_TESTS_BRANCH = sh(returnStdout: true, script: "cat apiBranch").trim()
                            def GIT_COMMIT_HASH = sh(returnStdout: true, script: "cat apiCommitSha").trim()
                            runAPItests(IMAGE, API_TESTS_BRANCH, GIT_COMMIT_HASH, CLIENT_URL, OWNER)
                        }
                    }
                }
                stage('Test: PMM-Testsuite') {
                    steps {
                        script {
                            unstash 'IMAGE'
                            unstash 'pmmQABranch'
                            unstash 'pmmQACommitSha'
                            def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                            def CLIENT_IMAGE = sh(returnStdout: true, script: "cat results/docker/CLIENT_TAG").trim()
                            def OWNER = sh(returnStdout: true, script: "cat OWNER").trim()
                            def CLIENT_URL = sh(returnStdout: true, script: "cat CLIENT_URL").trim()
                            def PMM_QA_GIT_BRANCH = sh(returnStdout: true, script: "cat pmmQABranch").trim()
                            def PMM_QA_GIT_COMMIT_HASH = sh(returnStdout: true, script: "cat pmmQACommitSha").trim()
                            runTestSuite(IMAGE, CLIENT_URL, PMM_QA_GIT_BRANCH, PMM_QA_GIT_COMMIT_HASH)
                        }
                    }
                }
                stage('Test: UI') {
                    steps {
                        script {
                            unstash 'IMAGE'
                            unstash 'pmmUITestBranch'
                            unstash 'pmmUITestsCommitSha'
                            def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                            def CLIENT_IMAGE = sh(returnStdout: true, script: "cat results/docker/CLIENT_TAG").trim()
                            def OWNER = sh(returnStdout: true, script: "cat OWNER").trim()
                            def CLIENT_URL = sh(returnStdout: true, script: "cat CLIENT_URL").trim()
                            def PMM_QA_GIT_BRANCH = sh(returnStdout: true, script: "cat pmmUITestBranch").trim()
                            def PMM_QA_GIT_COMMIT_HASH = sh(returnStdout: true, script: "cat pmmUITestsCommitSha").trim()
                            runUItests(IMAGE, CLIENT_URL, PMM_QA_GIT_BRANCH, PMM_QA_GIT_COMMIT_HASH)
                        }
                    }
                }
            }
        }
    }
    post {
        always {
            script {
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    if (env.CHANGE_URL) {
                        unstash 'IMAGE'
                        def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                        slackSend channel: '#pmm-ci', color: '#00FF00', message: "[${JOB_NAME}]: build finished - ${IMAGE}"
                    }
                } else {
                    slackSend channel: '#pmm-ci', color: '#FF0000', message: "[${JOB_NAME}]: build ${currentBuild.result}"
                }
            }
            sh 'sudo make clean'
            deleteDir()
        }
    }
}
