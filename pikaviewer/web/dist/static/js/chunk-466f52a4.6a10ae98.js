(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-466f52a4"],{"24d73":function(t,e,n){"use strict";n.d(e,"e",(function(){return o})),n.d(e,"d",(function(){return i})),n.d(e,"f",(function(){return r})),n.d(e,"a",(function(){return s})),n.d(e,"b",(function(){return u})),n.d(e,"c",(function(){return l}));var c=n("7424"),a=n("b775");function o(t){return Object(a["e"])(c["PLAYER_INFO"],a["a"].GET,{name:t})}function i(t){return Object(a["e"])(c["MAIL_INFO"],a["a"].GET,{uid:t})}function r(t){return Object(a["e"])(c["MAIL_SEND"],a["a"].POST,t)}function s(t){return Object(a["e"])(c["MAIL_DEL"],a["a"].POST,t)}function u(){return Object(a["e"])(c["MAIL_ENTIRE"],a["a"].GET)}function l(t,e){return Object(a["e"])(c["PLAYER_KICK_OFF"],a["a"].GET,{uid:t,sec:e})}},7424:function(t,e){var n="/v1",c="http://127.0.0.1:8888";t.exports={SYSTEM_NAME:"".concat(n,"/name"),LOGIN:"".concat(n,"/admin/user/login"),ROUTES:"".concat(n,"/routes"),NODES_MODULE:"".concat(n,"/node/module"),NODES_PIKA:"".concat(n,"/node/pika"),QUERY_DATA:"".concat(n,"/query/data"),All_API:"".concat(c,"/all"),HISTORY:"".concat(c,"/history"),REQUEST_API:"".concat(c,"/"),PID:"".concat(n,"/restart/pid"),RESTART:"".concat(n,"/restart/do"),LOG:"".concat(n,"/restart/log"),SYNC_MAP:"".concat(n,"/restart/sync/map"),OVERWRITE_DEV:"".concat(n,"/restart/overwrite/dev"),PLAYER_INFO:"".concat(n,"/admin/player/info"),MAIL_INFO:"".concat(n,"/admin/player/mail/info"),MAIL_SEND:"".concat(n,"/admin/player/mail/send"),MAIL_DEL:"".concat(n,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(n,"/admin/player/mail/entire"),PLAYER_KICK_OFF:"".concat(n,"/admin/player/kick/off"),ANNOUNCEMENT_INFO:"".concat(n,"/announcement"),ANNOUNCEMENT_LIST:"".concat(n,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(n,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(n,"/announcement/del"),BATTLE_LOG_FILES:"".concat(n,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(n,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(n,"/battle/log/download"),BUILD_DO:"".concat(n,"/build/do"),BUILD_LOG:"".concat(n,"/build/log"),VERSION_INFO:"".concat(n,"/version/info"),VERSION_SAVE:"".concat(n,"/version/save"),VERSION_UPLOAD:"".concat(n,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(n,"/version/zip"),BROADCAST_SEND:"".concat(n,"/broadcast/send"),WL_LIST:"".concat(n,"/whitelist/list"),WL_SAVE:"".concat(n,"/whitelist/save"),WL_DEL:"".concat(n,"/whitelist/del"),GITLAB_MEMBERS:"".concat(n,"/gitlab/members"),GITLAB_MODIFY:"".concat(n,"/gitlab/modify"),CDKEY_SAVE:"".concat(n,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(n,"/cdkey/deactive"),BWL_LIST:"".concat(n,"/beta/whitelist/list"),BWL_SAVE:"".concat(n,"/beta/whitelist/save"),BWL_DEL:"".concat(n,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(n,"/map/line/info")}},"7c23":function(t,e,n){},9992:function(t,e,n){"use strict";n.r(e);var c=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("a-card",[n("a-space",[n("a-input",{attrs:{"addon-before":"UID",placeholder:"请输入玩家的UID"},model:{value:t.uid,callback:function(e){t.uid=e},expression:"uid"}}),n("a-input",{staticStyle:{width:"270px"},attrs:{"addon-before":"秒",placeholder:" 请输入延迟下线时间（秒）\n    "},model:{value:t.sec,callback:function(e){t.sec=e},expression:"sec"}},[n("a-tooltip",{attrs:{slot:"suffix",title:"X秒后玩家被踢下线"},slot:"suffix"},[n("a-icon",{staticStyle:{color:"rgba(0,0,0,.45)"},attrs:{type:"question-circle"}})],1)],1),n("a-button",{attrs:{type:"primary",loading:t.loading},on:{click:t.send}},[t._v("确定")])],1)],1)},a=[],o=(n("96cf"),n("1da1")),i=n("24d73"),r={name:"Player",data:function(){return{uid:void 0,sec:void 0,loading:!1}},created:function(){},computed:{},methods:{onChange:function(t){this.type=t.target.value},send:function(){if(isNaN(this.uid)||this.uid<=0)this.$message.error("无效的UID");else if(isNaN(this.sec)||this.sec<0)this.$message.error("无效延迟时间");else{var t=this;this.$confirm({title:"提示",content:"确定将该玩家踢下线吗?",confirmButtonText:"确定",cancelButtonText:"取消",type:"warning",onOk:function(){return Object(o["a"])(regeneratorRuntime.mark((function e(){return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return e.next=2,t.doSend();case 2:case"end":return e.stop()}}),e)})))()},onCancel:function(){t.$message.info("已取消")}})}},doSend:function(){var t=this;return Object(o["a"])(regeneratorRuntime.mark((function e(){var n,c,a,o;return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return e.prev=0,t.loading=!0,e.next=4,Object(i["c"])(+t.uid,+t.sec);case 4:if(n=e.sent,c=n.data,a=c.code,o=c.msg,0===a){e.next=11;break}return t.$message.error(o),t.loading=!1,e.abrupt("return");case 11:t.$message.success(o),t.loading=!1,e.next=19;break;case 15:e.prev=15,e.t0=e["catch"](0),t.loading=!1,t.$message.error(e.t0.message);case 19:case"end":return e.stop()}}),e,null,[[0,15]])})))()}}},s=r,u=(n("ac5c"),n("0c7c")),l=Object(u["a"])(s,c,a,!1,null,"2f9695dc",null);e["default"]=l.exports},ac5c:function(t,e,n){"use strict";var c=n("7c23"),a=n.n(c);a.a}}]);