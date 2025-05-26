<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <title>{{.SITE_NAME}} - Powered by MinDoc</title>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="renderer" content="webkit">
    <meta name="author" content="Minho" />
    <meta name="site" content="https://www.iminho.me" />
    <meta name="keywords" content="MinDoc,文档在线管理系统,WIKI,wiki,wiki在线,文档在线管理,接口文档在线管理,接口文档管理">
    <meta name="description" content="MinDoc文档在线管理系统 {{.site_description}}">
    <!-- Bootstrap -->
    <link href="{{cdncss "/static/bootstrap/css/bootstrap.min.css"}}" rel="stylesheet">
    <link href="{{cdncss "/static/font-awesome/css/font-awesome.min.css"}}" rel="stylesheet">
    <link href="{{cdncss "/static/css/main.css" "version"}}" rel="stylesheet">
    <style>
        .label-section {
            margin-bottom: 30px;
            padding-bottom: 10px;
            border-bottom: 1px solid #eee;
        }
        .label-section h3 {
            margin-bottom: 20px;
            font-weight: 600;
        }
        .label-section .more-link {
            float: right;
            font-size: 14px;
            margin-top: 10px;
        }
        .view-mode {
            text-align: right;
            margin-bottom: 15px;
        }
        .view-mode a {
            margin-left: 15px;
            color: #666;
        }
        .view-mode a.active {
            color: #1abc9c;
            font-weight: bold;
        }
    </style>
</head>
<body>
<div class="manual-reader manual-container">
    {{template "widgets/header.tpl" .}}
    <div class="container manual-body">
        <div class="row">
            <div class="manual-list">
                <div class="view-mode">
                    <span>查看方式：</span>
                    <a href="{{urlfor "HomeController.Index"}}">全部</a>
                    <a href="?by_label=true" class="active">按标签</a>
                </div>

                {{range $labelName, $books := .LabelBooks}}
                <div class="label-section">
                    <h3>
                        <i class="fa fa-tag"></i> {{$labelName}}
                        <a href="{{urlfor "LabelController.Index" ":key" $labelName}}" class="more-link">更多 <i class="fa fa-angle-right"></i></a>
                    </h3>
                    {{range $index, $item := $books}}
                        <div class="list-item" data-id="{{$item.BookId}}">
                            <dl class="manual-item-standard">
                                <dt>
                                    <a href="{{urlfor "DocumentController.Index" ":key" $item.Identify}}" title="{{$item.BookName}}-{{$item.CreateName}}">
                                        <img src="{{cdnimg $item.Cover}}" class="cover" alt="{{$item.BookName}}-{{$item.CreateName}}" onerror="this.src='{{cdnimg "static/images/book.jpg"}}';">
                                    </a>
                                </dt>
                                <dd>
                                    <a href="{{urlfor "DocumentController.Index" ":key" $item.Identify}}" class="name" title="{{$item.BookName}}-{{$item.CreateName}}">{{$item.BookName}}</a>
                                </dd>
                                <dd>
                                <span class="author">
                                    <b class="text">{{i18n $.Lang "blog.author"}}</b>
                                    <b class="text">-</b>
                                    <b class="text">{{if eq $item.RealName "" }}{{$item.CreateName}}{{else}}{{$item.RealName}}{{end}}</b>
                                </span>
                                </dd>
                            </dl>
                        </div>
                    {{end}}
                    <div class="clearfix"></div>
                </div>
                {{else}}
                <div class="text-center" style="height: 200px;margin: 100px;font-size: 28px;">{{i18n $.Lang "message.no_project"}}</div>
                {{end}}
            </div>
        </div>
    </div>
    {{template "widgets/footer.tpl" .}}
</div>
<script src="{{cdnjs "/static/jquery/1.12.4/jquery.min.js"}}" type="text/javascript"></script>
<script src="{{cdnjs "/static/bootstrap/js/bootstrap.min.js"}}" type="text/javascript"></script>
<script src="{{cdnjs "/static/layer/layer.js"}}"></script>
</body>
</html>
