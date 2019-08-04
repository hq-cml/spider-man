![标题](./img/spider-title.png)

## 一个简单的爬虫框架
Spider-Man是一款基于Go实现的小型爬虫框架，支持从一个给定URL开始，递归爬取页面内容。  
目前已经和小型搜索引擎[Spider-Engine](https://github.com/hq-cml/spider-engine)打通，支持爬取内容灌入搜索引擎进一步分析。

#### 特点说明：
- 1. 利用了Go的高并发的特性, 高效控制单机的爬取并发度
- 2. 框架式的插件设计，用户只需要实现接口，即可自己定义Dom元素的分析逻辑
- 3. 框架式的插件设计，用户只需要实现接口，即可自己定义分析结果的处理逻辑，比如导入搜索引擎或者其他存储
- 4. 支持实时查看爬虫状态统计


#### 安装：
1. go get github.com/hq-cml/spider-man
2. cd 项目目录
3. go build ./

#### 插件说明：
目前提供了两个插件Demo：
- baseSpider：支持比较简单的关键字匹配功能, 比如从360主站https://www.360.cn开始搜索，打印出全部包含"老周"的网页  

```
运行：
./spider-man -c "conf/spider.conf" -f "https://www.360.cn" -u "老周"
./spider-man -c "conf/spider.conf" -f 'http://www.sohu.com' -u "张朝阳"
```

- engineSpider：实现了和搜索引擎Spider-Engine打通，爬取到的结果直接导入搜索引擎
```
运行：
./spider-man -c "conf/spider.conf" -p engine -f 'http://www.360.cn/news.html' -u '127.0.0.1:9528'
```

#### 查看运行状态

```
curl http://ip:8080/runInfo
```


#### 目录说明：
1. basic：基本数据类型定义
2. conf：配置文件
3. helper：业务无关的工具
4. logic: 核心业务代码
5. middleware: 中间件
6. plugin: 爬虫逻辑插件
7. vendor: 依赖

#### 架构设计：
![架构](./img/spider-struct.png)

说明
1. 调度器负责全局各个模块的调度, 核心工作是将请求从缓存中运送到请求Channel
2. downloader负责下载工作,产出是Response,下载并发度通过downloader池子来控制
3. analyzer负责分析工作,输入Response,产出是新的Request和Item项
4. processor负责最终Item的处理
5. 其中analyzer和processor的行为支持用户通过插件的形式定制
6. 中间件主要负责各个模块之间的缓冲

#### TODO：
1. 爬虫插件的逻辑丰富，能适应更多的页面风格
2. Cookie等带登陆的功能
3. 分布式爬虫
4. 防封禁
