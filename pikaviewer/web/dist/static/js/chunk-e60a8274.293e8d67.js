(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-e60a8274"],{1377:function(t,e,n){"use strict";var r=n("e25c"),a=n.n(r);a.a},7424:function(t,e){var n="https://10.23.20.53:9991/v1",r="http://127.0.0.1:8888";t.exports={LOGIN:"".concat(n,"/admin/user/login"),ROUTES:"".concat(n,"/routes"),NODES_MODULE:"".concat(n,"/node/module"),NODES_PIKA:"".concat(n,"/node/pika"),QUERY_DATA:"".concat(n,"/query/data"),All_API:"".concat(r,"/all"),HISTORY:"".concat(r,"/history"),REQUEST_API:"".concat(r,"/"),PID:"".concat(n,"/restart/pid"),RESTART:"".concat(n,"/restart/do"),LOG:"".concat(n,"/restart/log"),PLAYER_INFO:"".concat(n,"/admin/player/info"),MAIL_INFO:"".concat(n,"/admin/player/mail/info"),MAIL_SEND:"".concat(n,"/admin/player/mail/send"),MAIL_DEL:"".concat(n,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(n,"/admin/player/mail/entire"),PLAYER_KICK_OFF:"".concat(n,"/admin/player/kick/off"),ANNOUNCEMENT_INFO:"".concat(n,"/announcement"),ANNOUNCEMENT_LIST:"".concat(n,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(n,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(n,"/announcement/del"),BATTLE_LOG_FILES:"".concat(n,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(n,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(n,"/battle/log/download"),BUILD_DO:"".concat(n,"/build/do"),BUILD_LOG:"".concat(n,"/build/log"),VERSION_INFO:"".concat(n,"/version/info"),VERSION_SAVE:"".concat(n,"/version/save"),VERSION_UPLOAD:"".concat(n,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(n,"/version/zip"),BROADCAST_SEND:"".concat(n,"/broadcast/send"),WL_LIST:"".concat(n,"/whitelist/list"),WL_SAVE:"".concat(n,"/whitelist/save"),WL_DEL:"".concat(n,"/whitelist/del"),GITLAB_MEMBERS:"".concat(n,"/gitlab/members"),GITLAB_MODIFY:"".concat(n,"/gitlab/modify"),CDKEY_SAVE:"".concat(n,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(n,"/cdkey/deactive"),BWL_LIST:"".concat(n,"/beta/whitelist/list"),BWL_SAVE:"".concat(n,"/beta/whitelist/save"),BWL_DEL:"".concat(n,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(n,"/map/line/info")}},b624:function(t,e,n){"use strict";n.r(e);var r=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticStyle:{margin:"50px auto auto 200px"}},[n("h1",[t._v("战斗服PID："+t._s(t.battlePid))]),n("h1",[t._v("逻辑服PID："+t._s(t.logicPid))]),n("h1",[t._v("匹配服PID："+t._s(t.dungeonMatchPid))]),n("a-button",{staticStyle:{width:"220px",height:"64px","margin-top":"20px","font-size":"25px","font-weight":"bold"},attrs:{type:"danger",icon:"reload",loading:t.loading},on:{click:function(e){return t.restart(t.battle)}}},[t._v("重启战斗 ")]),n("a-button",{staticStyle:{width:"220px",height:"64px","margin-top":"20px","margin-left":"20px","font-size":"25px","font-weight":"bold"},attrs:{type:"danger",icon:"reload",loading:t.loading},on:{click:function(e){return t.restart(t.logic)}}},[t._v("重启逻辑服 ")]),n("a-button",{staticStyle:{width:"220px",height:"64px","margin-top":"20px","margin-left":"20px","font-size":"25px","font-weight":"bold"},attrs:{type:"danger",icon:"reload",loading:t.loading},on:{click:function(e){return t.restart(t.dungeonMatch)}}},[t._v("重启匹配服 ")]),n("div",{staticStyle:{"margin-top":"20px",overflow:"auto",height:"600px"},attrs:{id:"log"}},t._l(t.log,(function(e,r){return n("div",{key:r,staticStyle:{"line-height":"20px"}},[t._v(t._s(e))])})),0)],1)},a=[],c=(n("c975"),n("a15b"),n("96cf"),n("1da1")),o=n("7424"),i=n("b775");function s(t){return Object(i["e"])(o["PID"],i["a"].GET,{type:t})}function u(t){return Object(i["e"])(o["RESTART"],i["a"].GET,{type:t})}function l(t){return Object(i["e"])(o["LOG"],i["a"].GET,{type:t})}var d={name:"Restart",data:function(){return{battle:"1",logic:"2",dungeon:"3",dungeonMatch:"4",battlePid:"N/A",logicPid:"N/A",dungeonPid:"N/A",dungeonMatchPid:"N/A",loading:!1,log:[],logInterval:void 0}},created:function(){this.logInterval&&clearInterval(this.logInterval),this.fetchPid()},mounted:function(){return Object(c["a"])(regeneratorRuntime.mark((function t(){return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:case"end":return t.stop()}}),t)})))()},beforeDestroy:function(){this.logInterval&&clearInterval(this.logInterval)},computed:{},methods:{fetchPid:function(){var t=this;return Object(c["a"])(regeneratorRuntime.mark((function e(){var n,r,a,c,o,i,u,l;return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return e.prev=0,e.next=3,s();case 3:if(n=e.sent,r=n.status,a=n.data,200===r){e.next=8;break}return e.abrupt("return");case 8:0===a.code&&(c=a.data,o=c.battle,i=c.logic,u=c.dungeon,l=c.dungeonMatch,t.battlePid=o,t.logicPid=i,t.dungeonPid=u,t.dungeonMatchPid=l),e.next=13;break;case 11:e.prev=11,e.t0=e["catch"](0);case 13:case"end":return e.stop()}}),e,null,[[0,11]])})))()},restart:function(t){var e=this,n="";n="1"===t?"战斗服务器":"2"===t?"逻辑服务器":"3"===t?"副本服务器":"匹配服务器",this.$confirm({title:"提示",content:"确定要重启"+n+"吗?",confirmButtonText:"确定",cancelButtonText:"取消",type:"warning",onOk:function(){return Object(c["a"])(regeneratorRuntime.mark((function n(){var r;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return e.loading=!0,n.next=3,e.doRestart(t);case 3:if(r=n.sent,!r){n.next=10;break}return n.next=7,e.fetchLog(t);case 7:e.intervalFetchLog(t),n.next=11;break;case 10:e.loading=!1;case 11:case"end":return n.stop()}}),n)})))()},onCancel:function(){e.$message.info("已取消")}})},doRestart:function(t){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function n(){var r,a,c,o,i;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return n.prev=0,n.next=3,u(t);case 3:if(r=n.sent,a=r.status,c=r.data,200===a){n.next=9;break}return e.$message.error("重启发生错误："+a),n.abrupt("return",!1);case 9:if(o=c.code,i=c.msg,0===o){n.next=13;break}return e.$message.error("重启发生错误："+i),n.abrupt("return",!1);case 13:return n.abrupt("return",!0);case 16:return n.prev=16,n.t0=n["catch"](0),e.$message.error("重启发生错误："+n.t0.message),n.abrupt("return",!1);case 20:case"end":return n.stop()}}),n,null,[[0,16]])})))()},fetchLog:function(t){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function n(){var r,a,c,o,i,s;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return n.next=2,l(t);case 2:if(r=n.sent,a=r.status,c=r.data,200===a){n.next=8;break}return e.$message.error("请求错误"),n.abrupt("return");case 8:if(o=c.code,i=c.msg,0===o){n.next=12;break}return e.$message.error("获取日志发生错误："+i),n.abrupt("return");case 12:e.log=c.data,s=document.getElementById("log"),s.scrollTop=s.scrollHeight,e.doneCheck(t);case 16:case"end":return n.stop()}}),n)})))()},intervalFetchLog:function(t){var e=this;this.logInterval&&clearInterval(this.logInterval),this.stopFetchLogCheckCount=0,this.logInterval=setInterval(Object(c["a"])(regeneratorRuntime.mark((function n(){return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return n.next=2,e.fetchLog(t);case 2:case"end":return n.stop()}}),n)}))),2e3)},doneCheck:function(t){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function n(){var r,a,c;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:if(r=e.log.join(""),a=r.indexOf("restart error")>-1,!(r.indexOf("restart done")>-1||a)){n.next=9;break}return clearInterval(e.logInterval),e.loading=!1,n.next=7,e.fetchPid();case 7:c=e.getPid(t),a||"'N/A"===c?e.$error({title:"提示",content:"重启失败，请检查日志"}):e.$success({title:"提示",content:"重启成功"});case 9:case"end":return n.stop()}}),n)})))()},getPid:function(t){switch(t){case this.battle:return this.battlePid;case this.logic:return this.logicPid;case this.dungeon:return this.dungeonPid;case this.dungeonMatch:return this.dungeonMatchPid}return"N/A"}}},g=d,h=(n("1377"),n("0c7c")),p=Object(h["a"])(g,r,a,!1,null,"5cdbd7be",null);e["default"]=p.exports},e25c:function(t,e,n){}}]);