# ReadExcel
使用go读取Excel数据并生成对应文件

## 编译
```bash
cd src && go build -o ../example/bin/ReadExcel_linux
```

## 使用
在可执行文件目录下准备config.json和要导出表格目录, 填好配置执行
```json
{
    "input_path": "./tabby", // 表格目录
    "client": {
        "output_path": "./data" // 输出目录
    }
}
```

## 规则定义
1. 第一行文件设置
    1. 文件类型(横/竖, 默认横) SheetFileType [0/1]
    1. 文件名(不包含后缀, 默认为sheet名) OutFileName
    1. 主键数量 KeyNum
1. 第二行为第一行对应的值
1. 第三行字段名
1. 第四行字段类型(必须存在,支持?可为空,默认必须有值) 基础类型 bool int double string
1. 第五行字段分组(c/s, 默认全部)
1. '##'开头 表示注释 任意多行
1. 出现全空的行则丢弃空行后面的所有行, 除定义和注释行外

## 数据
1. 类型可为空时, 空格子为默认值

## TODO
1. 解决生成字段因map无序, 与表格中顺序不一样
1. 支持更多类型的数据文件格式和更多使用语言
