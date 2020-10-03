library changelog: false, identifier: 'lib@master', retriever: modernSCM([
    $class: 'GitSCMSource',
    remote: 'https://github.com/Percona-Lab/jenkins-pipelines.git'
]) _


void runStaging(String DOCKER_VERSION, CLIENT_VERSION) {
    stagingJob = build job: 'aws-staging-start', parameters: [
        string(name: 'DOCKER_VERSION', value: DOCKER_VERSION),
        string(name: 'CLIENT_VERSION', value: CLIENT_VERSION),
        string(name: 'PS_VERSION', value: '5.6'),
        string(name: 'CLIENTS', value: '--addclient=ps,1'),
        string(name: 'DOCKER_ENV_VARIABLE', value: '-e PMM_DEBUG=1 -e PERCONA_TEST_CHECKS_INTERVAL=10s -e PERCONA_TEST_DBAAS=1 -e PERCONA_TEST_AUTH_HOST=check-dev.percona.com:443 -e PERCONA_TEST_CHECKS_HOST=check-dev.percona.com:443 -e PERCONA_TEST_CHECKS_PUBLIC_KEY=RWTg+ZmCCjt7O8eWeAmTLAqW+1ozUbpRSKSwNTmO+exlS5KEIPYWuYdX'),
        string(name: 'NOTIFY', value: 'false'),
        string(name: 'DAYS', value: '1')
    ]
    env.VM_IP = stagingJob.buildVariables.IP
    env.VM_NAME = stagingJob.buildVariables.VM_NAME
}


void destroyStaging(IP) {
    build job: 'aws-staging-stop', parameters: [
        string(name: 'VM', value: IP),
    ]
}

void runAPItests(String DOCKER_IMAGE_VERSION, BRANCH_NAME, GIT_COMMIT_HASH, CLIENT_VERSION, OWNER, PMM_SERVER_IP) {
    def apiTestJob = build job: 'pmm2-api-tests-temp', wait: true, propagate: false, parameters: [
        string(name: 'DOCKER_VERSION', value: DOCKER_IMAGE_VERSION),
        string(name: 'GIT_BRANCH', value: BRANCH_NAME),
        string(name: 'OWNER', value: OWNER),
        string(name: 'GIT_COMMIT_HASH', value: GIT_COMMIT_HASH),
        string(name: 'SERVER_IP', value: PMM_SERVER_IP)
    ]
    env.API_TESTS_URL = apiTestJob.buildVariables.JOB_RUN_URL
    env.API_TESTS_RESULT = apiTestJob.result
}

void runTestSuite(String DOCKER_IMAGE_VERSION, CLIENT_VERSION, PMM_QA_GIT_BRANCH, PMM_QA_GIT_COMMIT_HASH, PMM_SERVER_IP) {
    testSuiteJob = build job: 'pmm2-testsuite-temp', parameters: [
        string(name: 'DOCKER_VERSION', value: DOCKER_IMAGE_VERSION),
        string(name: 'CLIENT_VERSION', value: CLIENT_VERSION),
        string(name: 'PMM_QA_GIT_BRANCH', value: PMM_QA_GIT_BRANCH),
        string(name: 'PMM_QA_GIT_COMMIT_HASH', value: PMM_QA_GIT_COMMIT_HASH),
        string(name: 'SERVER_IP', value: PMM_SERVER_IP)
    ]
    env.testSuiteJobUrl = testSuiteJob.buildVariables.BUILD_URL
}

void runUItests(String DOCKER_IMAGE_VERSION, CLIENT_VERSION, PMM_QA_GIT_BRANCH, PMM_QA_GIT_COMMIT_HASH, PMM_SERVER_IP) {
    e2eTestJob = build job: 'pmm2-ui-tests', parameters: [
        string(name: 'DOCKER_VERSION', value: DOCKER_IMAGE_VERSION),
        string(name: 'CLIENT_VERSION', value: CLIENT_VERSION),
        string(name: 'GIT_BRANCH', value: PMM_QA_GIT_BRANCH),
        string(name: 'GIT_COMMIT_HASH', value: PMM_QA_GIT_COMMIT_HASH),
        string(name: 'SERVER_IP', value: PMM_SERVER_IP),
        string(name: 'CLIENT_INSTANCE', value: 'yes')
    ]
    env.e2eTestJobUrl = e2eTestJob.buildVariables.BUILD_URL
}

void addComment(String COMMENT) {
    withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
        sh """
            curl -v -X POST \
                -H "Authorization: token ${GITHUB_API_TOKEN}" \
                -d "{\\"body\\":\\"${COMMENT}"
        """
    }
}

def isBranchBuild = true
def apiTestsFailed = false
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
                    set -o errexit
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
        stage('Execute Tests') {
            when {
                expression {
                    !isBranchBuild
                }
            }
            parallel {
                stage('Test: API') {
                    steps {
                        script {
                            unstash 'apiBranch'
                            unstash 'apiCommitSha'
                            def OWNER = sh(returnStdout: true, script: "cat OWNER").trim()
                            def API_TESTS_BRANCH = sh(returnStdout: true, script: "cat apiBranch").trim()
                            def GIT_COMMIT_HASH = sh(returnStdout: true, script: "cat apiCommitSha").trim()
                            runAPItests('dev-latest', API_TESTS_BRANCH, GIT_COMMIT_HASH, 'dev-latest', OWNER, '3.137.218.245')
                            apiTestsFailed = true
                        }
                    }
                }
            }
        }
    }
    post {
        always {
            script {
                if (env.VM_IP) {
                    destroyStaging(env.VM_IP)
                }
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    if (env.CHANGE_URL) {
                        unstash 'IMAGE'
                        def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                        slackSend channel: '#pmm-ci', color: '#00FF00', message: "[${JOB_NAME}]: build finished - ${IMAGE}"
                    }
                } else {
                    sh 'echo ${API_TESTS_URL}'
                    if(env.API_TESTS_RESULT == "FAILURE")
                    {
                        addComment("Link to Failed API tests Job: ${API_TESTS_URL}")
                    }
                    slackSend channel: '#pmm-ci', color: '#FF0000', message: "[${JOB_NAME}]: build ${currentBuild.result}"
                }
            }
            sh 'sudo make clean'
            deleteDir()
        }
    }
}
