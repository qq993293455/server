(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-4bbcd9c1"],{"030a":function(e,t,n){"use strict";n.r(t);var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticStyle:{margin:"50px auto auto 100px"}},[n("span",{staticStyle:{"font-weight":"bold","font-size":"20px"}},[e._v("请选择要打包的分支：")]),n("a-select",{staticStyle:{width:"120px"},attrs:{"default-value":e.branch,value:e.branch,size:"large"},on:{change:e.handleChange}},[n("a-select-option",{attrs:{value:"develop"}},[e._v(" develop ")]),n("a-select-option",{attrs:{value:"patch"}},[e._v(" patch ")])],1),n("a-button",{staticStyle:{"margin-top":"20px","margin-left":"5px"},attrs:{type:"primary",size:"large",icon:"tool",loading:e.loading},on:{click:e.build}},[e._v("打包 ")]),n("div",{staticStyle:{"margin-top":"20px",overflow:"auto",height:"600px"},attrs:{id:"log"}},e._l(e.log,(function(t,a){return n("div",{key:a,staticStyle:{"line-height":"20px"}},[e._v(e._s(t))])})),0)],1)},r=[],c=(n("c975"),n("a15b"),n("96cf"),n("1da1")),o=n("7424"),i=n("b775");function s(e){return Object(i["e"])(o["BUILD_DO"],i["a"].POST,{branch:e})}function l(){return Object(i["e"])(o["BUILD_LOG"],i["a"].GET)}var u={name:"Build",data:function(){return{loading:!1,log:[],logInterval:void 0,branch:"develop"}},created:function(){this.logInterval&&clearInterval(this.logInterval)},mounted:function(){return Object(c["a"])(regeneratorRuntime.mark((function e(){return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:case"end":return e.stop()}}),e)})))()},beforeDestroy:function(){this.logInterval&&clearInterval(this.logInterval)},computed:{},methods:{build:function(){var e=this;this.$confirm({title:"提示",content:"该操作可能耗时较长，确定打包 "+e.branch+" 分支吗?",confirmButtonText:"确定",cancelButtonText:"取消",type:"warning",onOk:function(){return Object(c["a"])(regeneratorRuntime.mark((function t(){var n;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return e.loading=!0,t.next=3,e.doBuild();case 3:if(n=t.sent,!n){t.next=10;break}return t.next=7,e.fetchLog();case 7:e.intervalFetchLog(),t.next=11;break;case 10:e.loading=!1;case 11:case"end":return t.stop()}}),t)})))()},onCancel:function(){e.$message.info("已取消")}})},doBuild:function(){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function t(){var n,a,r,c,o;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return t.prev=0,t.next=3,s(e.branch);case 3:if(n=t.sent,a=n.status,r=n.data,200===a){t.next=9;break}return e.$message.error("打包出错："+a),t.abrupt("return",!1);case 9:if(c=r.code,o=r.msg,0===c){t.next=13;break}return e.$message.error("打包出错："+o),t.abrupt("return",!1);case 13:return t.abrupt("return",!0);case 16:return t.prev=16,t.t0=t["catch"](0),e.$message.error("打包出错："+t.t0.message),t.abrupt("return",!1);case 20:case"end":return t.stop()}}),t,null,[[0,16]])})))()},fetchLog:function(e){var t=this;return Object(c["a"])(regeneratorRuntime.mark((function n(){var a,r,c,o,i,s;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return n.next=2,l(e);case 2:if(a=n.sent,r=a.status,c=a.data,200===r){n.next=8;break}return t.$message.error("请求错误"),n.abrupt("return");case 8:if(o=c.code,i=c.msg,0===o){n.next=12;break}return t.$message.error("获取日志发生错误："+i),n.abrupt("return");case 12:t.log=c.data,s=document.getElementById("log"),s.scrollTop=s.scrollHeight+100,t.doneCheck();case 16:case"end":return n.stop()}}),n)})))()},intervalFetchLog:function(e){var t=this;this.logInterval&&clearInterval(this.logInterval),this.stopFetchLogCheckCount=0,this.logInterval=setInterval(Object(c["a"])(regeneratorRuntime.mark((function n(){return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return n.next=2,t.fetchLog(e);case 2:case"end":return n.stop()}}),n)}))),2e3)},doneCheck:function(){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function t(){var n;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:n=e.log.join(""),n.indexOf("succ")>-1&&(clearInterval(e.logInterval),e.loading=!1,e.$success({title:"提示",content:"打包完成"}));case 2:case"end":return t.stop()}}),t)})))()},handleChange:function(e){this.branch=e}}},d=u,g=(n("ca18"),n("0c7c")),h=Object(g["a"])(d,a,r,!1,null,"291585e5",null);t["default"]=h.exports},7424:function(e,t){var n="v1",a="http://127.0.0.1:8888";e.exports={SYSTEM_NAME:"".concat(n,"/name"),LOGIN:"".concat(n,"/admin/user/login"),ROUTES:"".concat(n,"/routes"),NODES_MODULE:"".concat(n,"/node/module"),NODES_PIKA:"".concat(n,"/node/pika"),QUERY_DATA:"".concat(n,"/query/data"),All_API:"".concat(a,"/all"),HISTORY:"".concat(a,"/history"),REQUEST_API:"".concat(a,"/"),PID:"".concat(n,"/restart/pid"),RESTART:"".concat(n,"/restart/do"),LOG:"".concat(n,"/restart/log"),SYNC_MAP:"".concat(n,"/restart/sync/map"),PLAYER_INFO:"".concat(n,"/admin/player/info"),MAIL_INFO:"".concat(n,"/admin/player/mail/info"),MAIL_SEND:"".concat(n,"/admin/player/mail/send"),MAIL_DEL:"".concat(n,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(n,"/admin/player/mail/entire"),PLAYER_KICK_OFF:"".concat(n,"/admin/player/kick/off"),ANNOUNCEMENT_INFO:"".concat(n,"/announcement"),ANNOUNCEMENT_LIST:"".concat(n,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(n,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(n,"/announcement/del"),BATTLE_LOG_FILES:"".concat(n,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(n,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(n,"/battle/log/download"),BUILD_DO:"".concat(n,"/build/do"),BUILD_LOG:"".concat(n,"/build/log"),VERSION_INFO:"".concat(n,"/version/info"),VERSION_SAVE:"".concat(n,"/version/save"),VERSION_UPLOAD:"".concat(n,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(n,"/version/zip"),BROADCAST_SEND:"".concat(n,"/broadcast/send"),WL_LIST:"".concat(n,"/whitelist/list"),WL_SAVE:"".concat(n,"/whitelist/save"),WL_DEL:"".concat(n,"/whitelist/del"),GITLAB_MEMBERS:"".concat(n,"/gitlab/members"),GITLAB_MODIFY:"".concat(n,"/gitlab/modify"),CDKEY_SAVE:"".concat(n,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(n,"/cdkey/deactive"),BWL_LIST:"".concat(n,"/beta/whitelist/list"),BWL_SAVE:"".concat(n,"/beta/whitelist/save"),BWL_DEL:"".concat(n,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(n,"/map/line/info")}},ca18:function(e,t,n){"use strict";var a=n("f1a1"),r=n.n(a);r.a},f1a1:function(e,t,n){}}]);