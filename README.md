![标题](./img/spider-title.png)

## 一个简单的爬虫框架
t
#### 说明：
1. 一个简单的爬虫框架，支持从一个给定URL开始，分析出新的URL递归爬并对当前页面内容分析
2. 利用了Go的高并发的特性, 高效控制单机的爬去并发度
3. 利用接口封装, 保证可扩展, 比如中间件和爬取插件等等,都可以在实现接口的前提下进行替换
4. 目前只实现了一个基本插件, 支持比较简单的关键字匹配功能, 比如从360主站开始搜索全部"老周"的网页

#### 安装运行：
1. go get github.com/hq-cml/spider-man
2. cd 项目目录
3. go build ./
4. 运行：
    -    ./spider-man -c "conf/spider.conf" -f "https://www.360.cn" -u "老周"
    -    ./spider-man -c "conf/spider.conf" -f 'http://www.sohu.com' -u "张朝阳"
5. 查看运行状态
    -  curl http://ip:8080/runInfo

#### 目录说明：
1. basic：基本数据类型定义
2. conf：配置文件
3. helper：业务无关的工具
4. logic: 核心业务代码
5. middleware: 中间件
6. plugin: 爬虫逻辑插件
7. vendor: 依赖

#### 架构：
![架构](./img/spider-struct.png)

如上图
1. 调度器负责全局各个模块的调度, 核心工作是将请求从缓存中运送到请求Channel
2. downloader负责下载工作,产出是Response,下载并发度通过downloader池子来控制
3. analyzer负责分析工作,输入Response,产出是新的Request和Item项
4. processor负责最终Item的处理
5. 其中analyzer和processor的行为支持用户通过插件的形式定制
6. 中间件主要负责各个模块之间的缓冲

#### TODO：
1. 爬虫插件的逻辑丰富
2. Cookie等带登陆的功能
3. 分布式爬虫
4. 防封禁
5. 爬取内容结构化存储，和SE打通