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
    <!-- Bootstrap -->    <link href="{{cdncss "/static/bootstrap/css/bootstrap.min.css"}}" rel="stylesheet">
    <link href="{{cdncss "/static/font-awesome/css/font-awesome.min.css"}}" rel="stylesheet">
    <link href="{{cdncss "/static/css/main.css" "version"}}" rel="stylesheet">
    <link href="{{cdncss "/static/css/itemspace.css" "version"}}" rel="stylesheet">
    <script type="text/javascript">
        window.updateBookOrder = "{{urlfor "BookController.UpdateBookOrder"}}";
    </script>
</head>
<body>
<div class="manual-reader manual-container">
    {{template "widgets/header.tpl" .}}
    <div class="container manual-body">
        <div class="row">
             <div class="manual-list">
                {{range $idx, $itemId := .GroupedOrder}}
                    {{$books := index $.GroupedBooks $itemId}}
                    {{if gt (len $books) 0}}
                        {{if eq $itemId 0}}
                            <div class="panel panel-default">
                                <div class="panel-heading">
                                    <h3 class="panel-title">
                                        <i class="fa fa-book"></i> 未分组项目
                                        <span class="badge">{{len $books}}</span>
                                    </h3>
                                </div>
                                <div class="panel-body">
                        {{else}}
                            {{$itemset := index $.ItemsetsMap $itemId}}
                            <div class="panel panel-default">
                                <div class="panel-heading">
                                    <h3 class="panel-title">
                                        <i class="fa fa-folder-open"></i> {{$itemset.ItemName}}
                                        <span class="badge">{{len $books}}</span>
                                    </h3>
                                    {{if ne $itemset.Description ""}}
                                        <p class="text-muted small">{{$itemset.Description}}</p>
                                    {{end}}
                                </div>
                                <div class="panel-body">
                        {{end}}

                        {{range $index,$item := $books}}
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
                        </div>
                    {{end}}
                {{end}}
                {{if eq .TotalPages 0}}
                    <div class="text-center" style="height: 200px;margin: 100px;font-size: 28px;">{{i18n $.Lang "message.no_project"}}</div>
                {{end}}
            </div>
            <nav class="pagination-container">
                {{if gt .TotalPages 1}}
                    {{.PageHtml}}
                {{end}}
                <div class="clearfix"></div>
            </nav>
        </div>
    </div>
    {{template "widgets/footer.tpl" .}}
</div>
<script src="{{cdnjs "/static/jquery/1.12.4/jquery.min.js"}}" type="text/javascript"></script>
<script src="{{cdnjs "/static/bootstrap/js/bootstrap.min.js"}}" type="text/javascript"></script>
<script src="{{cdnjs "/static/layer/layer.js"}}"></script>
<script src="{{cdnjs "/static/js/sort.js"}}" type="text/javascript"></script>
{{.Scripts}}
</body>
</html>