(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-7597c1d4"],{"24d73":function(e,t,a){"use strict";a.d(t,"d",(function(){return c})),a.d(t,"c",(function(){return o})),a.d(t,"e",(function(){return i})),a.d(t,"a",(function(){return s})),a.d(t,"b",(function(){return l}));var n=a("7424"),r=a("b775");function c(e){return Object(r["e"])(n["PLAYER_INFO"],r["a"].GET,{name:e})}function o(e){return Object(r["e"])(n["MAIL_INFO"],r["a"].GET,{uid:e})}function i(e){return Object(r["e"])(n["MAIL_SEND"],r["a"].POST,e)}function s(e){return Object(r["e"])(n["MAIL_DEL"],r["a"].POST,e)}function l(){return Object(r["e"])(n["MAIL_ENTIRE"],r["a"].GET)}},"498a":function(e,t,a){"use strict";var n=a("23e7"),r=a("58a8").trim,c=a("c8d2");n({target:"String",proto:!0,forced:c("trim")},{trim:function(){return r(this)}})},5899:function(e,t){e.exports="\t\n\v\f\r                　\u2028\u2029\ufeff"},"58a8":function(e,t,a){var n=a("1d80"),r=a("5899"),c="["+r+"]",o=RegExp("^"+c+c+"*"),i=RegExp(c+c+"*$"),s=function(e){return function(t){var a=String(n(t));return 1&e&&(a=a.replace(o,"")),2&e&&(a=a.replace(i,"")),a}};e.exports={start:s(1),end:s(2),trim:s(3)}},7424:function(e,t){var a="https://10.23.20.53:9991/v1",n="http://127.0.0.1:8888";e.exports={LOGIN:"".concat(a,"/admin/user/login"),ROUTES:"".concat(a,"/routes"),NODES_MODULE:"".concat(a,"/node/module"),NODES_PIKA:"".concat(a,"/node/pika"),QUERY_DATA:"".concat(a,"/query/data"),All_API:"".concat(n,"/all"),HISTORY:"".concat(n,"/history"),REQUEST_API:"".concat(n,"/"),PID:"".concat(a,"/restart/pid"),RESTART:"".concat(a,"/restart/do"),LOG:"".concat(a,"/restart/log"),PLAYER_INFO:"".concat(a,"/admin/player/info"),MAIL_INFO:"".concat(a,"/admin/player/mail/info"),MAIL_SEND:"".concat(a,"/admin/player/mail/send"),MAIL_DEL:"".concat(a,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(a,"/admin/player/mail/entire"),ANNOUNCEMENT_INFO:"".concat(a,"/announcement"),ANNOUNCEMENT_LIST:"".concat(a,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(a,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(a,"/announcement/del"),BATTLE_LOG_FILES:"".concat(a,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(a,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(a,"/battle/log/download"),BUILD_DO:"".concat(a,"/build/do"),BUILD_LOG:"".concat(a,"/build/log"),VERSION_INFO:"".concat(a,"/version/info"),VERSION_SAVE:"".concat(a,"/version/save"),VERSION_UPLOAD:"".concat(a,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(a,"/version/zip"),BROADCAST_SEND:"".concat(a,"/broadcast/send"),WL_LIST:"".concat(a,"/whitelist/list"),WL_SAVE:"".concat(a,"/whitelist/save"),WL_DEL:"".concat(a,"/whitelist/del"),GITLAB_MEMBERS:"".concat(a,"/gitlab/members"),GITLAB_MODIFY:"".concat(a,"/gitlab/modify"),CDKEY_SAVE:"".concat(a,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(a,"/cdkey/deactive"),BWL_LIST:"".concat(a,"/beta/whitelist/list"),BWL_SAVE:"".concat(a,"/beta/whitelist/save"),BWL_DEL:"".concat(a,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(a,"/map/line/info")}},c3fc8:function(e,t,a){"use strict";a.r(t);var n=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticStyle:{margin:"50px"}},[a("a-form-model",{ref:"form",attrs:{layout:"inline",rules:e.rules,model:e.form},on:{submit:e.handleSubmit},nativeOn:{submit:function(e){e.preventDefault()}}},[a("a-form-model-item",{attrs:{label:""}},[a("a-radio-group",{attrs:{"default-value":e.form.way,value:e.form.way,"button-style":"solid"},on:{change:e.wayChange}},[a("a-radio-button",{attrs:{value:"name"}},[e._v(" 通过玩家名查询 ")]),a("a-radio-button",{attrs:{value:"uid"}},[e._v(" 通过UID查询 ")]),a("a-radio-button",{attrs:{value:"entire"}},[e._v(" 全服邮件 ")])],1)],1),"entire"!==e.form.way?a("a-form-model-item",{attrs:{prop:"keywords"}},[a("a-input",{attrs:{allowClear:"",placeholder:"关键字"},model:{value:e.form.keywords,callback:function(t){e.$set(e.form,"keywords",t)},expression:"form.keywords"}},[a("a-icon",{staticStyle:{color:"rgba(0,0,0,.25)"}})],1)],1):e._e(),a("a-form-model-item",[a("a-button",{attrs:{type:"primary","html-type":"submit",icon:"search",loading:e.loading}},[e._v(" 查询 ")])],1)],1),e.total>0?a("a-card",[a("div",{staticStyle:{"margin-bottom":"20px","font-size":"15px"}},[e.total>0?[e._v(" 总共找到"),a("span",{staticStyle:{"font-size":"20px","font-weight":"bold",color:"red"}},[e._v(" "+e._s(e.total))]),e._v(" 个玩家 "),e.total>5?[e._v(" ，仅显示前"),a("span",{staticStyle:{"font-size":"20px","font-weight":"bold",color:"red"}},[e._v(" 5 ")]),e._v("条 ")]:e._e()]:e._e()],2),a("a-table",{attrs:{columns:e.playerColumns,"data-source":e.playerList,pagination:!1,size:"small"},scopedSlots:e._u([{key:"uid",fn:function(t){return a("span",{},[a("a",{staticClass:"ant-dropdown-link",on:{click:function(a){return e.selectPlayer(t)}}},[e._v(" "+e._s(t)+" ")])])}}],null,!1,1655047331)})],1):e._e(),a("a-card",[e.mails.length>0?[a("div",{staticStyle:{"margin-bottom":"16px"}},[a("a-button",{attrs:{type:"danger",icon:"delete",disabled:!e.hasSelected,loading:e.loading},on:{click:e.deleteMail}},[e._v(" 删除 ")]),a("span",{staticStyle:{"margin-left":"8px"}},[e.hasSelected?[e._v(" "+e._s("选中 "+e.selectedRowKeys.length+" 封邮件")+" ")]:e._e()],2)],1),a("a-table",{attrs:{columns:e.columns,"data-source":e.mails,pagination:!1,size:"small","row-selection":{selectedRowKeys:e.selectedRowKeys,onChange:e.onSelectChange}},scopedSlots:e._u([{key:"read",fn:function(t){return a("span",{},[a("a-tag",{attrs:{color:t?"geekblue":"green"}},[e._v(" "+e._s(t?"已读":"未读")+" ")])],1)}}],null,!1,1315798911)})]:e._e()],2)],1)},r=[],c=(a("4160"),a("498a"),a("159b"),a("96cf"),a("1da1")),o=a("24d73"),i=[{title:"邮件ID",dataIndex:"id"},{title:"邮件类型",dataIndex:"type"},{title:"发送者",dataIndex:"sender"},{title:"文本ID",dataIndex:"text_id"},{title:"标题",dataIndex:"title"},{title:"内容",dataIndex:"content"},{title:"hi",dataIndex:"hi"},{title:"生效时间",dataIndex:"activated_at"},{title:"过期时间",dataIndex:"expired_at"},{title:"参数",dataIndex:"args"},{title:"附件",dataIndex:"attachment"},{title:"是否已读",dataIndex:"read",scopedSlots:{customRender:"read"}},{title:"创建时间",dataIndex:"created_at"}],s=[{title:"UID",dataIndex:"uid",scopedSlots:{customRender:"uid"}},{title:"昵称",dataIndex:"nickname"},{title:"创建时间",dataIndex:"create_time"},{title:"最后一次登录时间",dataIndex:"login_time"}],l={name:"MailQuery",components:{},data:function(){return{columns:i,playerColumns:s,mails:[],loading:!1,form:{way:"name",keywords:""},total:0,playerList:[],selectedRowKeys:[],rules:{keywords:[{required:!0,message:"请输入要查询的关键字",trigger:"blur"},{min:1,max:16,message:"长度1-16",trigger:"blur"}]}}},created:function(){},mounted:function(){},computed:{hasSelected:function(){return this.selectedRowKeys.length>0}},methods:{onSelectChange:function(e){this.selectedRowKeys=e},query:function(){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function t(){var a;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:if(a=e.form.keywords.trim(),"name"!==e.form.way){t.next=6;break}return t.next=4,e.fetchPlayerInfo(a);case 4:t.next=17;break;case 6:if("uid"!==e.form.way){t.next=11;break}return t.next=9,e.fetchMailInfo(a);case 9:t.next=17;break;case 11:if("entire"!==e.form.way){t.next=16;break}return t.next=14,e.fetchEntireMail();case 14:t.next=17;break;case 16:e.$message.error("无效的查询方式");case 17:case"end":return t.stop()}}),t)})))()},fetchPlayerInfo:function(e){var t=this;return Object(c["a"])(regeneratorRuntime.mark((function a(){var n,r,c,i,s,l,u;return regeneratorRuntime.wrap((function(a){while(1)switch(a.prev=a.next){case 0:return a.prev=0,a.next=3,Object(o["d"])(e);case 3:if(n=a.sent,r=n.data,c=r.code,i=r.msg,0===c){a.next=9;break}return t.$message.error(i),a.abrupt("return");case 9:s=r.data,l=s.list,u=s.total,t.total=u,l=l||[],l.sort((function(e,t){return t.login_time-e.login_time})),l.forEach((function(e){e.key=e.uid,e.create_time=t.$convertTime(new Date(e.create_time),"YYYY-MM-DD HH:mm:ss"),e.login_time=t.$convertTime(new Date(e.login_time),"YYYY-MM-DD HH:mm:ss")})),t.playerList=l,a.next=20;break;case 17:a.prev=17,a.t0=a["catch"](0),t.$message.error(a.t0.message);case 20:case"end":return a.stop()}}),a,null,[[0,17]])})))()},fetchMailInfo:function(e){var t=this;return Object(c["a"])(regeneratorRuntime.mark((function a(){var n,r,c,i,s;return regeneratorRuntime.wrap((function(a){while(1)switch(a.prev=a.next){case 0:return t.mails=[],a.prev=1,a.next=4,Object(o["c"])(e);case 4:if(n=a.sent,r=n.data,c=r.code,i=r.msg,0===c){a.next=10;break}return t.$message.error(i),a.abrupt("return");case 10:s=r.data||[],s.sort((function(e,t){return t.created_at-e.created_at})),s.forEach((function(e){e.key=e.id,e.activated_at=t.$convertTime(new Date(e.activated_at),"YYYY-MM-DD HH:mm:ss"),e.expired_at=t.$convertTime(new Date(e.expired_at),"YYYY-MM-DD HH:mm:ss"),e.created_at=t.$convertTime(new Date(e.created_at),"YYYY-MM-DD HH:mm:ss"),e.args&&e.args.length>0&&(e.args=JSON.stringify(e.args)),e.attachment&&e.attachment.length>0&&(e.attachment=JSON.stringify(e.attachment))})),t.mails=s||[],t.mails.length<=0&&t.$message.info("当前玩家没有邮件"),a.next=20;break;case 17:a.prev=17,a.t0=a["catch"](1),t.$message.error(a.t0.message);case 20:case"end":return a.stop()}}),a,null,[[1,17]])})))()},wayChange:function(e){this.form.way=e.target.value},handleSubmit:function(){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function t(){return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return e.mails=[],e.playerList=[],e.total=0,t.prev=3,t.next=6,e.$refs.form.validate();case 6:t.next=11;break;case 8:return t.prev=8,t.t0=t["catch"](3),t.abrupt("return");case 11:return e.loading=!0,t.next=14,e.query();case 14:e.loading=!1,"name"===e.form.way&&e.total<=0&&e.$message.info("未找到对应的玩家");case 16:case"end":return t.stop()}}),t,null,[[3,8]])})))()},deleteMail:function(){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function t(){var a;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:a=e,e.$confirm({title:"警告",content:"删除后无法恢复，确定删除这些邮件吗？",okType:"danger",onOk:function(){return Object(c["a"])(regeneratorRuntime.mark((function e(){var t,n,r,i;return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return a.loading=!0,e.prev=1,e.next=4,Object(o["a"])({entire:"entire"===a.form.way,uid:+a.form.keywords,mailId:a.selectedRowKeys});case 4:if(t=e.sent,n=t.data,r=n.code,i=n.msg,0===r){e.next=11;break}return a.loading=!1,a.$message.error(i),e.abrupt("return");case 11:a.selectedRowKeys=[],setTimeout(Object(c["a"])(regeneratorRuntime.mark((function e(){return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return e.next=2,a.query();case 2:a.loading=!1,a.$message.success("删除成功");case 4:case"end":return e.stop()}}),e)}))),1e3),e.next=19;break;case 15:e.prev=15,e.t0=e["catch"](1),a.loading=!1,a.$message.error(e.t0.message);case 19:case"end":return e.stop()}}),e,null,[[1,15]])})))()}});case 2:case"end":return t.stop()}}),t)})))()},selectPlayer:function(e){var t=this;return Object(c["a"])(regeneratorRuntime.mark((function a(){return regeneratorRuntime.wrap((function(a){while(1)switch(a.prev=a.next){case 0:return t.form.way="uid",t.form.keywords=e+"",t.total=0,a.next=5,t.fetchMailInfo(e);case 5:case"end":return a.stop()}}),a)})))()},fetchEntireMail:function(){var e=this;return Object(c["a"])(regeneratorRuntime.mark((function t(){var a,n,r,c,i;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return e.mails=[],t.prev=1,t.next=4,Object(o["b"])();case 4:if(a=t.sent,n=a.data,r=n.code,c=n.msg,0===r){t.next=10;break}return e.$message.error(c),t.abrupt("return");case 10:i=n.data||[],i.sort((function(e,t){return t.created_at-e.created_at})),i.forEach((function(t){t.key=t.id,t.activated_at=e.$convertTime(new Date(t.activated_at),"YYYY-MM-DD HH:mm:ss"),t.expired_at=e.$convertTime(new Date(t.expired_at),"YYYY-MM-DD HH:mm:ss"),t.created_at=e.$convertTime(new Date(t.created_at),"YYYY-MM-DD HH:mm:ss"),t.args&&t.args.length>0&&(t.args=JSON.stringify(t.args)),t.attachment&&t.attachment.length>0&&(t.attachment=JSON.stringify(t.attachment))})),e.mails=i||[],e.mails.length<=0&&e.$message.info("没有全服邮件"),t.next=20;break;case 17:t.prev=17,t.t0=t["catch"](1),e.$message.error(t.t0.message);case 20:case"end":return t.stop()}}),t,null,[[1,17]])})))()}}},u=l,d=a("0c7c"),m=Object(d["a"])(u,n,r,!1,null,"61bf5be8",null);t["default"]=m.exports},c8d2:function(e,t,a){var n=a("d039"),r=a("5899"),c="​᠎";e.exports=function(e){return n((function(){return!!r[e]()||c[e]()!=c||r[e].name!==e}))}}}]);