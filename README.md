持续集成流程
---

1. 开发团队提交代码到gitlab
2. gitlab检测到master分支代码发生提交
3. gitlab使用webhook触发jenkins启动持续集成
4. jenkins从gitlab拉取代码，并自动配置编译环境
5. jenkins进行go项目自动代码测试
6. jenkins进行go项目构建打包
7. jenkins调用远程docker，把go项目编译包打包成docker镜像
8. jenkins调用远程docker，把打包的docker镜像push到远程registry
9. jenkins调用远程kubernetes，通知kubernetes从registry拉取最新的docker镜像，并且通过滚动更新的方式，把应用发布到生产环境，当应用完全发布后，下线旧的服务，中间不会中断服务。


使用GoDep
---
* 安装godep
```sh
go get github.com/tools/godep
```

* Godep需要golang.org/x/sys/unix包
```sh
go get golang.org/x/sys/unix
```

* 进入程序目录，使用godep save保存相关依赖到Godeps
```sh
cd /Volumes/Share/Go_program/src/godemo
godep save -v ./...
```

* 成功后会自动创建./Godeps目录，下边有Godeps.json配置文件，记录相关依赖的以及依赖版本；同时会自动创建./vendor，并且把依赖包复制到本目录下

* 进入程序目录，使用godep restore获取相关依赖
```sh
cd /Volumes/Share/Go_program/src/godemo
godep restore -v
```

* 运行go程序
```sh
godep go run gin/gin_demo.go
```

* 编译go程序
```sh
godep go build gin/gin_demo.go
```


Jenkins与GitLab、Docker、Registry、GoLang的集成
---

### Jenkins安装相关插件（"系统管理" -> "管理插件"）

- CloudBees Docker Build and Publish plugin
    > docker插件，在"构建"步骤增加"Docker Build and Publish"，把构建结果Build到docker以及push到registry
- CloudBees Docker Custom Build Environment Plugin
    > docker插件，在"构建环境"步骤增加"Build inside a Docker container"，在构建环境的时候下载docker客户端，在docker中进行项目构建
- docker-build-step
    > docker插件，在"构建"步骤增加"Execute Docker command"，在构建过程中增加docker客户端指令步骤
- Go Plugin
    > golang插件，在"构建环境"步骤增加"Set up Go programming language tools"，在构建环境的时候下载golang环境
- GitLab Plugin
    > gitlab插件，在"General"步骤增加"GitLab connection"，源码管理可以调用gitlab
- Gitlab Authentication plugin
    > gitlab插件，可以使用gitlab的api token进行授权
- Gitlab Hook Plugin
    > gitlab插件，在"构建触发器"步骤增加"Build when a change is pushed to GitLab. GitLab CI Service URL: http://localhost:8080/project/XXX"
    
    > 当gitlab代码发生提交的时候，通过gitlab hook主动触发构建 
- Kubernetes plugin
    > kubernetes插件，可以在kubernetes中启动相关pod
    
### Jenkins中GitLab、Docker以及GoLang基础配置

- GitLab连接设置（"系统管理" -> "系统设置" -> "GitLab connections"）
    > "Connection name" 设置为 gitlab_cookeem
    
    > "Gitlab host URL" 设置为 http://gitlab
     
    > "Credentials" 需要"Add Credentials"，"Kind" 选择 "GitLab API token"；"API token"对应 Gitlab "User Settings" -> "Account" -> "Private token"
    
    > "Test Connection" 检测GitLab API token能够正常连接

- Docker环境设置（"系统管理" -> "Global Tool Configuration" -> "Docker" -> "Docker安装"）
    > "新增Docker" 新增一个Docker版本的环境变量
    
    > "Name" 设置为 docker_1.13.1；"自动安装" 选择上
    
    > "新增安装" 选择 "Install latest from docker.io"
    
    > "Docker version" 设置为 1.13.1
    
- Docker Builder环境设置，对应docker-build-step插件（"系统管理" -> "系统设置" -> "Docker Builder"）
    > "Docker URL" 设置为 tcp://docker:2375
    
    > "Test Connection" 检测连接是否正常
    
- GoLang环境设置（"系统管理" -> "Global Tool Configuration" -> "Go" -> "Go安装"）
    > "新增Go" 新增一个Go版本的环境变量
    
    > "别名" 设置为 go_1.8.1；"自动安装" 选择上
    
    > "新增安装" 选择 "Install from golang.org"
    
    > "Version" 选择 GoLang的版本

### Jenkins中新建项目，实现GoLang项目通过GitLab进行源码管理和自动打包到Docker

- "新建" -> "构建一个自由风格的软件项目"

- "General"设置
    > "项目名称" 设置为 godemo
    
    > "GitLab connection" 选择 gitlab_cookeem（对应"系统管理" -> "系统设置" -> "GitLab connections"）

- "源码管理"设置
    > "Git" -> "Repositories" -> "Repository URL" 设置为 http://gitlab/cookeem/godemo
    
    > "Git" -> "Repositories" -> "Credentials" -> "Add Credentials"，"Kind" 选择 "Username with password"，"Username" 设置为 cookeem@qq.com，"Password" 设置为对应GitLab账号密码

- "构建触发器"设置
    > "Build when a change is pushed to GitLab. GitLab CI Service URL: http://localhost:8080/project/godemo" 该项选择
    
    > "Build when a change is pushed to GitLab." -> "高级" -> "Secret token" -> "Generate" 创建Jenkins token
    
    > 打开GitLab界面，"Projects" -> "cookeem/godemo" -> "Settings" -> "Integrations"，"URL" 设置为 http://jenkins:8080/project/godemo（对应Jenkins的"GitLab CI Service URL"），"Secret Token" 设置为对应Jenkins的"Secret token"。创建WebHook后进行测试，就会触发自动构建

- "构建环境"设置
    > "Add timestamps to the Console Output" 选择上
    
    > "Set up Go programming language tools" -> "Go version" 选择 go_1.8.1
    
- "构建"设置
    > "新增构建步骤" -> "Execute shell"，执行以下构建脚本
```
# pwd
# echo "###################"
# ls -al
# echo "###################"
# 设置GOPATH，保证go的编译
export GOPATH=`pwd`
# 设置DOCKER_HOME
export MY_DOCKER_HOME=/var/jenkins_home/tools/org.jenkinsci.plugins.docker.commons.tools.DockerTool/docker_1.13.1
export PATH=$PATH:$MY_DOCKER_HOME/bin
# 从VERSION文件获取APP版本信息
export APP_VERSION_NAME=`cat VERSION`
echo "###################"
# printenv
# 设置go依赖包的位置为src
rm -rf src
mv vendor src
echo "###################"
# 编译前测试
go test ./test
echo "###################"
echo "###################"
echo "# begin go testing"
# 进行go编译打包
go build -ldflags "-X main.VersionName=$APP_VERSION_NAME" gin/gin_demo.go
echo "# end of go testing"
echo "###################"
echo "###################"
# 把go编译包打包成docker镜像
docker -H tcp://ci-docker:2375 build -t k8s-registry:5000/cookeem/godemo:$APP_VERSION_NAME .
docker -H tcp://ci-docker:2375 tag k8s-registry:5000/cookeem/godemo:$APP_VERSION_NAME k8s-registry:5000/cookeem/godemo:latest
# 把docker镜像推送到k8s-registry:5000
docker -H tcp://ci-docker:2375 push k8s-registry:5000/cookeem/godemo:$APP_VERSION_NAME
docker -H tcp://ci-docker:2375 push k8s-registry:5000/cookeem/godemo:latest
# 使用kubectl拉起镜像
kubectl apply -f deploy/kubernetes/godemo.yaml
```
    
    > "新增构建步骤" -> "Docker Build and Publish" （该项不要设置）
    ```
       "Repository Name" 设置为 godemo
       "Tag" 设置为 0.1.0 
       "Docker Host URI" 设置为 tcp://docker:2375 （连接远程docker）
       "Server credentials" 设置为 none
       "Docker registry URL" 设置为 http://registry:5000/v2/
       "Registry credentials" 设置为 none
       "Force Pull" 取消选择
       "Docker installation" 选择 docker_1.13.1，在构建的时候自动安装docker客户端
    ```

    > "新增构建步骤" -> "Execute shell"，执行以下构建脚本 （该项不要设置）
    ```
        /var/jenkins_home/tools/org.jenkinsci.plugins.docker.commons.tools.DockerTool/docker_1.13.1/bin/docker -H tcp://docker:2375 stop godemo
        /var/jenkins_home/tools/org.jenkinsci.plugins.docker.commons.tools.DockerTool/docker_1.13.1/bin/docker -H tcp://docker:2375 rm godemo
        /var/jenkins_home/tools/org.jenkinsci.plugins.docker.commons.tools.DockerTool/docker_1.13.1/bin/docker -H tcp://docker:2375 run -d --name godemo -p 8081:8081 registry:5000/godemo:latest
    ```

- "保存"项目

- GitLab中进行push，触发Jenkins进行GoLang项目构建，完成构建后，把编译包build成docker镜像，并且把镜像push到docker registry

- 源码的根目录需要创建Dockerfile，用于"CloudBees Docker Build and Publish plugin"进行自动构建docker镜像

- 在jenkins容器中测试godemo是否启动正常
    ```
        docker exec -ti jenkins bash
        curl docker:8081/user/haijian/ok
        exit
    ```

- 在docker容器中测试godemo是否启动正常，检测logs中的App Version
    ```
        docker exec -ti docker ash
        docker images
        docker ps
        docker logs godemo
        exit
    ```

- 关闭服务，注意，如果只是stop再up，docker容器启动会出现异常
    ```
        docker-compose stop && docker-compose rm -f
    ```

# 使用docker编译linux_amd64可运行包

- 因为MacOSX下边编译的包不能在linux下运行，因此需要使用linux进行go项目编译

- 在godemo目录下运行docker，并执行以下命令进行包编译 

    ```sh
    # 运行容器
    docker run -ti --rm -v "$PWD":/usr/src/godemo -w /usr/src/godemo golang:alpine ash
    
    # 以下命令在container中运行
    mkdir /lib64
    ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
    export GOPATH=$(pwd)
    mv vendor src
    go build gin/gin_demo.go
    mv src vendor
    ```

