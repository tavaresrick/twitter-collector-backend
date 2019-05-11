node() {
    def image = ''
    
    stage('Pull') {
        sh 'echo "Cleaning WorkDir"'
        cleanWs()
        sh 'echo "Fetching repo"'
        git credentialsId: 'Github', url: 'https://github.com/tavaresrick/twitter-collector-backend'
    }
    stage('Build docker image') {
        sh 'echo "Building docker image"'
        image = docker.build("tavaresrick/twitter-collector-backend:1.${env.BUILD_ID}")
        image.tag("latest")
    }
    stage('Unit test') {
        sh 'echo "Performing unit testing"'
        withCredentials([string(credentialsId: 'TWITTER_CONSUMER_KEY', variable: 'twConsumerKey'), string(credentialsId: 'TWITTER_CONSUMER_SECRET', variable: 'twConsumerSecret'), string(credentialsId: 'TWITTER_APP_ACCESS_TOKEN', variable: 'twAppAccessToken'), string(credentialsId: 'TWITTER_ACCESS_SECRET', variable: 'twAccessSecret')]) {
            withEnv(["TW_CONSUMER_KEY=${twConsumerKey}", "TW_CONSUMER_SECRET=${twConsumerSecret}", "TW_ACCESS_SECRET=${twAccessSecret}", "TW_ACCESS_TOKEN=${twAppAccessToken}", "TAG=1.${env.BUILD_ID}"]) {
                sh 'docker stack deploy -c docker-compose.yaml tc_build_test'
            }
        }
        sh 'sleep 20'
        sh 'docker exec $( docker ps --filter name=tc_build_test_backend* -q)  go test -v .'
        sh 'docker stack rm tc_build_test'
    }
    stage('Push image') {
        sh 'echo "Pushing docker image to registry"'
        withCredentials([usernamePassword(credentialsId: 'Dockerhub', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
          sh "sudo docker login -u $USERNAME -p '$PASSWORD'"
          image.push()
          image.push('latest')
        }
    }
    stage('Deploy') {
        sh 'echo "Deploying"'
        withCredentials([sshUserPrivateKey(credentialsId: 'SSHKeyDockerManager', keyFileVariable: 'KEYFILE', passphraseVariable: '', usernameVariable: 'USER')]) {
            def remote = [:]
            remote.name = 'manager'
            remote.host = '172.31.41.11'
            remote.user = USER
            remote.identityFile = KEYFILE
            remote.allowAnyHosts = true
            sshCommand remote: remote, sudo: true, command: "docker service update --image tavaresrick/twitter-collector-backend:1.${env.BUILD_ID} backend_backend"
        }
    }
}