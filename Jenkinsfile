library changelog: false, identifier: 'lib@master', retriever: modernSCM([
    $class: 'GitSCMSource',
    remote: 'https://github.com/Percona-Lab/jenkins-pipelines.git'
]) _

pipeline {
    agent {
        label 'min-centos-7-x64'
    }
    stages {
        stage('Prepare') {
            steps {
                sh 'rm -rf tmp results'
                installDocker()
                sh '''
                    sudo yum -y install centos-release-scl-rh
                    sudo yum -y install rh-git29
                    sudo yum -y remove git
                    sudo ln -fs /opt/rh/rh-git29/root/usr/bin/git /usr/bin/git
                    sudo ln -fs /opt/rh/httpd24/root/usr/lib64/libcurl-httpd24.so.4 /usr/lib64/libcurl-httpd24.so.4
                    sudo ln -fs /opt/rh/httpd24/root/usr/lib64/libnghttp2-httpd24.so.14 /usr/lib64/libnghttp2-httpd24.so.14
                '''
                sh '''
                    git submodule init
                    git submodule update --remote
                '''
                slackSend channel: '#pmm-ci', color: '#FFFF00', message: "[${JOB_NAME}]: build started - ${BUILD_URL}"
            }
        }
        stage('Build client source') {
            steps {
                sh 'sg docker -c "./build/bin/build-client-source"'
            }
        }
        stage('Build client binary') {
            steps {
                sh 'sg docker -c "./build/bin/build-client-binary"'
                archiveArtifacts 'results/binary/pmm-client-*.tar.gz'
            }
        }
        stage('Build server packages') {
            steps {
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
                        build-server-rpm grafana
                    "
                '''
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
                sh 'sg docker -c "SAVE_DOCKER=1 ./build/bin/build-server-docker"'
                archiveArtifacts 'results/docker/pmm-server-*.docker'
            }
        }
    }
    post {
        always {
            script {
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    if (env.CHANGE_URL) {
                        withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
                            sh """
                                set -o xtrace
                                curl -v -X POST \
                                    -H "Authorization: token ${GITHUB_API_TOKEN}" \
                                    -d "{\\"body\\":\\"docker - \$(cat results/docker/TAG)\\nclient - ${BUILD_URL}artifact/results/binary/\\"}" \
                                    "https://api.github.com/repos/\$(echo $CHANGE_URL | cut -d '/' -f 4-5)/issues/${CHANGE_ID}/comments"
                            """
                        }
                    }
                    slackSend channel: '#pmm-ci', color: '#00FF00', message: "[${JOB_NAME}]: build finished"
                } else {
                    slackSend channel: '#pmm-ci', color: '#FF0000', message: "[${JOB_NAME}]: build ${currentBuild.result}"
                }
            }
            deleteDir()
        }
    }
}
