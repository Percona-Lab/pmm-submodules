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
                    git submodule update
                '''
                slackSend channel: '#pmm-ci', color: '#FFFF00', message: "[${JOB_NAME}]: build started - ${BUILD_URL}"
            }
        }
        stage('Build client source') {
            steps {
                sh 'sg docker -c "./build/bin/build-pmm-client-source-tarball"'
            }
        }
        stage('Build client binary') {
            steps {
                sh 'sg docker -c "./build/bin/build-pmm-client-binary-tarball"'
                archiveArtifacts 'results/binary/pmm-client-*.tar.gz'
            }
        }
        stage('Build server packages') {
            steps {
                sh '''
                    sg docker -c "
                        export PATH=$PATH:$(pwd -P)/build/bin

                        # 1st-party
                        build-rpm percona-dashboards grafana-dashboards
                        build-rpm pmm-manage
                        build-rpm pmm-managed
                        build-rpm percona-qan-api qan-api
                        build-rpm percona-qan-app qan-app
                        build-rpm pmm-server
                        build-rpm pmm-update

                        # 3rd-party
                        build-rpm consul
                        build-rpm orchestrator
                        build-rpm rds_exporter
                        build-rpm prometheus
                        build-rpm grafana
                    "
                '''
            }
        }
        stage('Build server docker') {
            steps {
                sh 'sg docker -c "SAVE_DOCKER=1 ./build/bin/build-pmm-server-docker"'
                archiveArtifacts 'results/docker/pmm-server-*.docker'
            }
        }
    }
    post {
        always {
            script {
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    slackSend channel: '#pmm-ci', color: '#00FF00', message: "[${JOB_NAME}]: build finished"
                } else {
                    slackSend channel: '#pmm-ci', color: '#FF0000', message: "[${JOB_NAME}]: build ${currentBuild.result}"
                }
            }
            deleteDir()
        }
    }
}
