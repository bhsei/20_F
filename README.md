# 20_F_Gitea

## 小组成员

| 学号      | 姓名   | 
|:----------|:-------|
| SY1906915 | 麦梓健 | 
| SY1906505 | 孙维华 | 
| SY1906306 | 洪治凑 | 
| SY1906432 | 王子璇 | 
| SY1906117 | 王伟民 | 
| SY1906434 | 郑锋   | 

## gRPC生成

以下命令假定工作目录位于`3-项目源码/`。

将协议定义编译至Python代码用如下命令：
```
python -m grpc_tools.protoc -I. --python_out=./module_server --grpc_python_out=./module_server service.proto
```
相关依赖请使用 `pip install -r module_server/requirements.txt ` 安装

编译至Go代码用如下命令：
```
protoc --gogofast_out=plugins=grpc:./services/module service.proto
```
相关依赖安装请参考[gogo](https://github.com/gogo/protobuf)

## 通知服务模块

通知服务模块位于`3-项目源码/module_server`，入口文件为`mod_entry.py`，使用下列命令运行：
```
python mod_entry.py -r <root_path> -p <port>
```
可以使用`python mod_entry.py -h`查看帮助。

## 贡献

### 提交格式描述(Git Commit Message)

```
<类型>:<空格><简略描述>
<空行>
- <更改1说明>
	* <更改1.1说明>
- <更改2说明>
- <更改3说明>
...
```

在上述格式说明大致分为两个部分(使用空行分割):
 - 标题行: 必填, 描述主要修改类型和内容
 - 主题内容: 描述为什么修改, 做了什么样的修改, 以及开发的思路等等

若能通过简略描述说明提交内容，第二部分的内容可省略。

本项目管理中涉及到的对象见下表。

| 对象         | 包含内容                                        |
|:-------------|:------------------------------------------------|
| 文档         | 项目说明文档、代码文档、代码注释                |
| 代码         | 项目业务代码                                    |
| 测试用例     | 项目测试代码及附属文件                          |
| 项目配置文件 | 项目环境及git配置文件、组员合作说明文件         |
| 文件结构     | 项目文件的组织顺序                              |


对于项目中不同对象的更改，现暂定使用如下表所示的提交类型。

| 类型      | 适用对象     | 使用场景     |
|:----------|:-------------|:-------------|
| manage    | 项目配置文件 | 内容更改     |
| docs      | 文档         | 内容更新     |
| style     | 所有文件     | 格式修改     |
| refactor  | 代码         | 重构         |
| feature   | 代码         | 新功能特性   |
| fix       | 代码或文档   | 内容问题修正 |
| structure | 文件结构     | 修改目录编排 |
| test      | 测试用例     | 内容更改     |
