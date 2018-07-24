library changelog: false, identifier: 'lib@master', retriever: modernSCM([
    $class: 'GitSCMSource',
    remote: 'https://github.com/Percona-Lab/jenkins-pipelines.git'
]) _

pipeline {
    agent {
        label 'large-amazon'
    }
    stages {
        stage('Prepare') {
            steps {
                sh '''
                    sudo rm -rf results tmp || :
                    git reset --hard
                    git clean -fdx
                    git submodule foreach --recursive git reset --hard
                    git submodule foreach --recursive git clean -fdx
                '''
                installDocker()
                slackSend channel: '#pmm-ci', color: '#FFFF00', message: "[${JOB_NAME}]: build started - ${BUILD_URL}"
            }
        }
        stage('Build client source') {
            steps {
                sh '''
                    sg docker -c "
                        env
                        export pmm_version=$(cat VERSION)
                        ./build/bin/build-client-source
                    "
                '''
            }
        }
        stage('Build client binary') {
            steps {
                sh '''
                    sg docker -c "
                        export pmm_version=$(cat VERSION)
                        ./build/bin/build-client-binary
                    "
                '''
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
                    sh '''
                        aws s3 cp \
                            --acl public-read \
                            results/tarball/pmm-client-*.tar.gz \
                            s3://pmm-build-cache/pmm-client/pmm-client-${BRANCH_NAME}-${GIT_COMMIT:0:7}.tar.gz
                    '''
                }
            }
        }
        stage('Build client source rpm') {
            steps {
                sh 'sg docker -c "./build/bin/build-client-srpm centos:6"'
            }
        }
        stage('Build client binary rpm') {
            steps {
                sh '''
                    sg docker -c "
                        ./build/bin/build-client-rpm centos:7

                        mkdir -p tmp/pmm-server/RPMS/
                        cp results/rpm/pmm-client-*.rpm tmp/pmm-server/RPMS/
                    "
                '''
            }
        }
        stage('Build server packages') {
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
                    sh '''
                        sg docker -c "
                            export PATH=$PATH:$(pwd -P)/build/bin

                            # 1st-party
                            build-server-rpm percona-dashboards grafana-dashboards
                            build-server-rpm pmm-manage
                            build-server-rpm pmm-managed
                            build-server-rpm percona-qan-api qan-api
                            build-server-rpm percona-qan-app qan-app
                            build-server-rpm pmm-server
                            build-server-rpm pmm-update

                            # 3rd-party
                            build-server-rpm consul
                            build-server-rpm orchestrator
                            build-server-rpm rds_exporter
                            build-server-rpm prometheus
                            build-server-rpm prometheus1
                            build-server-rpm grafana
                        "
                    '''
                }
            }
        }
        stage('Build server docker') {
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
    }
    post {
        always {
            script {
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    unstash 'IMAGE'
                    def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                    if (env.CHANGE_URL) {
                        withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
                            sh """
                                set -o xtrace
                                curl -v -X POST \
                                    -H "Authorization: token ${GITHUB_API_TOKEN}" \
                                    -d "{\\"body\\":\\"docker - ${IMAGE}\\nclient - https://s3.us-east-2.amazonaws.com/pmm-build-cache/pmm-client/pmm-client-${BRANCH_NAME}-\${GIT_COMMIT:0:7}.tar.gz\\"}" \
                                    "https://api.github.com/repos/\$(echo $CHANGE_URL | cut -d '/' -f 4-5)/issues/${CHANGE_ID}/comments"
                            """
                        }
                    }
                    slackSend channel: '#pmm-ci', color: '#00FF00', message: "[${JOB_NAME}]: build finished - ${IMAGE}"
                } else {
                    slackSend channel: '#pmm-ci', color: '#FF0000', message: "[${JOB_NAME}]: build ${currentBuild.result}"
                }
            }
            sh 'sudo make clean'
            deleteDir()
        }
    }
}
