(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-7d19a41e"],{7424:function(e,t){var n="/v1",a="http://127.0.0.1:8888";e.exports={SYSTEM_NAME:"".concat(n,"/name"),LOGIN:"".concat(n,"/admin/user/login"),ROUTES:"".concat(n,"/routes"),NODES_MODULE:"".concat(n,"/node/module"),NODES_PIKA:"".concat(n,"/node/pika"),QUERY_DATA:"".concat(n,"/query/data"),All_API:"".concat(a,"/all"),HISTORY:"".concat(a,"/history"),REQUEST_API:"".concat(a,"/"),PID:"".concat(n,"/restart/pid"),RESTART:"".concat(n,"/restart/do"),LOG:"".concat(n,"/restart/log"),SYNC_MAP:"".concat(n,"/restart/sync/map"),OVERWRITE_DEV:"".concat(n,"/restart/overwrite/dev"),PLAYER_INFO:"".concat(n,"/admin/player/info"),MAIL_INFO:"".concat(n,"/admin/player/mail/info"),MAIL_SEND:"".concat(n,"/admin/player/mail/send"),MAIL_DEL:"".concat(n,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(n,"/admin/player/mail/entire"),PLAYER_KICK_OFF:"".concat(n,"/admin/player/kick/off"),ANNOUNCEMENT_INFO:"".concat(n,"/announcement"),ANNOUNCEMENT_LIST:"".concat(n,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(n,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(n,"/announcement/del"),BATTLE_LOG_FILES:"".concat(n,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(n,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(n,"/battle/log/download"),BUILD_DO:"".concat(n,"/build/do"),BUILD_LOG:"".concat(n,"/build/log"),VERSION_INFO:"".concat(n,"/version/info"),VERSION_SAVE:"".concat(n,"/version/save"),VERSION_UPLOAD:"".concat(n,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(n,"/version/zip"),BROADCAST_SEND:"".concat(n,"/broadcast/send"),WL_LIST:"".concat(n,"/whitelist/list"),WL_SAVE:"".concat(n,"/whitelist/save"),WL_DEL:"".concat(n,"/whitelist/del"),GITLAB_MEMBERS:"".concat(n,"/gitlab/members"),GITLAB_MODIFY:"".concat(n,"/gitlab/modify"),CDKEY_SAVE:"".concat(n,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(n,"/cdkey/deactive"),BWL_LIST:"".concat(n,"/beta/whitelist/list"),BWL_SAVE:"".concat(n,"/beta/whitelist/save"),BWL_DEL:"".concat(n,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(n,"/map/line/info")}},fde1:function(e,t,n){"use strict";n.r(t);var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("a-card",[e._v(" 项目： "),n("a-select",{staticStyle:{width:"80%"},attrs:{mode:"multiple",placeholder:"请选择项目（可多选）"},on:{change:e.projectChange}},e._l(e.projects,(function(t){return n("a-select-option",{key:t.id,attrs:{value:t.id}},[e._v(" "+e._s(t.name)+" ")])})),1),n("br"),n("br"),e._v(" 权限： "),n("a-select",{staticStyle:{width:"80%"},attrs:{placeholder:"请选择访问权限"},on:{change:e.accessLevelChange}},e._l(e.accessLevel,(function(t){return n("a-select-option",{key:t,attrs:{value:t}},[e._v(" "+e._s(t)+" ")])})),1),n("br"),n("br"),e._v(" 用户： "),n("a-select",{staticStyle:{width:"80%"},attrs:{mode:"multiple","filter-option":e.userFilter,placeholder:"请选择用户（可多选）"},on:{change:e.membersChange}},e._l(e.members,(function(t){return n("a-select-option",{key:t.id,attrs:{value:t.id}},[e._v(" "+e._s(t.username+" - "+t.name)+" ")])})),1),n("br"),n("br"),n("a-button",{staticStyle:{left:"50px"},attrs:{icon:"edit",type:"primary"},on:{click:e.modify}},[e._v("修改")])],1)},c=[],r=(n("c975"),n("96cf"),n("1da1")),o=n("7424"),s=n("b775");function i(){return Object(s["e"])(o["GITLAB_MEMBERS"],s["a"].GET)}function l(e){return Object(s["e"])(o["GITLAB_MODIFY"],s["a"].POST,e)}var u={name:"Gitlab",components:{},data:function(){return{projects:[{id:3,name:"coin-server"},{id:40,name:"l5client"},{id:7,name:"share"}],accessLevel:["Developer","Maintainer"],members:[],data:{members:[],projects:[],accessLevel:void 0}}},created:function(){this.fetchMembers()},methods:{projectChange:function(e){this.data.projects=e},accessLevelChange:function(e){this.data.accessLevel=e},membersChange:function(e){this.data.members=e},fetchMembers:function(){var e=this;return Object(r["a"])(regeneratorRuntime.mark((function t(){var n,a,c,r;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return t.prev=0,t.next=3,i();case 3:if(n=t.sent,a=n.data,c=a.code,r=a.msg,0===c){t.next=9;break}return e.$message.error(r),t.abrupt("return");case 9:e.members=a.data||[],t.next=15;break;case 12:t.prev=12,t.t0=t["catch"](0),e.$message.error(t.t0.message);case 15:case"end":return t.stop()}}),t,null,[[0,12]])})))()},modify:function(){var e=this;this.$confirm({title:"提示",content:"确定修改吗？",confirmButtonText:"确定",cancelButtonText:"取消",type:"warning",onOk:function(){return Object(r["a"])(regeneratorRuntime.mark((function t(){return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return t.next=2,e.doModify();case 2:case"end":return t.stop()}}),t)})))()}})},doModify:function(){var e=this;return Object(r["a"])(regeneratorRuntime.mark((function t(){var n,a,c,r;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return t.prev=0,t.next=3,l({access_level:e.data.accessLevel,projects:e.data.projects,users:e.data.members});case 3:if(n=t.sent,a=n.data,c=a.code,r=a.msg,0===c){t.next=8;break}return e.$message.error(r),t.abrupt("return");case 8:e.$message.success(r),t.next=14;break;case 11:t.prev=11,t.t0=t["catch"](0),e.$message.error(t.t0.message||t.t0);case 14:case"end":return t.stop()}}),t,null,[[0,11]])})))()},userFilter:function(e,t){return t.componentOptions.children[0].text.toLowerCase().indexOf(e.toLowerCase())>=0}}},d=u,m=n("0c7c"),_=Object(m["a"])(d,a,c,!1,null,"4c8a9d68",null);t["default"]=_.exports}}]);