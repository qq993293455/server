(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-8642da68"],{1301:function(e,t,a){},7424:function(e,t){var a="https://10.23.20.53:9991/v1",r="http://127.0.0.1:8888";e.exports={LOGIN:"".concat(a,"/admin/user/login"),ROUTES:"".concat(a,"/routes"),NODES_MODULE:"".concat(a,"/node/module"),NODES_PIKA:"".concat(a,"/node/pika"),QUERY_DATA:"".concat(a,"/query/data"),All_API:"".concat(r,"/all"),HISTORY:"".concat(r,"/history"),REQUEST_API:"".concat(r,"/"),PID:"".concat(a,"/restart/pid"),RESTART:"".concat(a,"/restart/do"),LOG:"".concat(a,"/restart/log"),PLAYER_INFO:"".concat(a,"/admin/player/info"),MAIL_INFO:"".concat(a,"/admin/player/mail/info"),MAIL_SEND:"".concat(a,"/admin/player/mail/send"),MAIL_DEL:"".concat(a,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(a,"/admin/player/mail/entire"),ANNOUNCEMENT_INFO:"".concat(a,"/announcement"),ANNOUNCEMENT_LIST:"".concat(a,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(a,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(a,"/announcement/del"),BATTLE_LOG_FILES:"".concat(a,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(a,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(a,"/battle/log/download"),BUILD_DO:"".concat(a,"/build/do"),BUILD_LOG:"".concat(a,"/build/log"),VERSION_INFO:"".concat(a,"/version/info"),VERSION_SAVE:"".concat(a,"/version/save"),VERSION_UPLOAD:"".concat(a,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(a,"/version/zip"),BROADCAST_SEND:"".concat(a,"/broadcast/send"),WL_LIST:"".concat(a,"/whitelist/list"),WL_SAVE:"".concat(a,"/whitelist/save"),WL_DEL:"".concat(a,"/whitelist/del"),GITLAB_MEMBERS:"".concat(a,"/gitlab/members"),GITLAB_MODIFY:"".concat(a,"/gitlab/modify"),CDKEY_SAVE:"".concat(a,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(a,"/cdkey/deactive"),BWL_LIST:"".concat(a,"/beta/whitelist/list"),BWL_SAVE:"".concat(a,"/beta/whitelist/save"),BWL_DEL:"".concat(a,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(a,"/map/line/info")}},"7adb":function(e,t,a){"use strict";var r=a("1301"),n=a.n(r);n.a},c2f1:function(e,t,a){"use strict";a.r(t);var r=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("a-card",[a("a-tabs",{attrs:{"default-active-key":"info"},on:{change:e.tabChange}},[a("a-tab-pane",{key:"info",attrs:{tab:"正常版本"}},[a("a-space",{attrs:{direction:"vertical",size:20}},[a("div",[e._v("主版本："+e._s(e.form.version))]),a("div",[e._v("最小版本："+e._s(e.form.min_version)+" "),a("a-tooltip",[a("template",{slot:"title"},[e._v(" 位于主版本和最小版本之间的版本均可进入游戏 ")]),a("a-icon",{staticStyle:{color:"#1890ff"},attrs:{type:"question-circle"}})],2)],1),a("div",[e._v("网关地址："+e._s(e.form.gateway))]),a("div",[e._v("热更文件地址："+e._s(e.form.cdn))]),a("div",[e._v("公告地址："+e._s(e.form.announcement))]),a("div",[e._v("版本文件名："+e._s(e.form.version_file))]),a("a-button",{attrs:{type:"primary"},on:{click:function(t){return e.modify(!1)}}},[e._v("修改")])],1)],1),a("a-tab-pane",{key:"audit",attrs:{tab:"送审版本"}},[a("a-space",{attrs:{direction:"vertical",size:20}},[a("div",[e._v("是否开启： "),a("a-switch",{attrs:{checked:e.activate},on:{change:e.switchChange}},[a("a-icon",{attrs:{slot:"checkedChildren",type:"check"},slot:"checkedChildren"}),a("a-icon",{attrs:{slot:"unCheckedChildren",type:"close"},slot:"unCheckedChildren"})],1)],1),a("div",[e._v("主版本："+e._s(e.form.version))]),a("div",[e._v("网关地址："+e._s(e.form.gateway))]),a("div",[e._v("热更文件地址："+e._s(e.form.cdn))]),a("div",[e._v("公告地址："+e._s(e.form.announcement))]),a("div",[e._v("版本文件名："+e._s(e.form.version_file))]),a("a-button",{attrs:{type:"primary"},on:{click:function(t){return e.modify(!0)}}},[e._v("修改")])],1)],1),a("a-tab-pane",{key:"version",attrs:{tab:"版本文件上传"}},[a("a-space",{staticStyle:{width:"100%"},attrs:{direction:"vertical"}},[a("a-upload-dragger",{attrs:{name:"file",multiple:!1,"default-file-list":e.defaultVersionFileList,"before-upload":e.beforeUpload}},[a("p",{staticClass:"ant-upload-drag-icon"},[a("a-icon",{attrs:{type:"inbox"}})],1),a("p",{staticClass:"ant-upload-text"},[e._v(" 单击或拖动文件到此区域上传 ")]),a("p",{staticClass:"ant-upload-hint"},[e._v(" 仅支持单个文件，且文件类型必须为.txt ")])]),a("a-row",{staticStyle:{"margin-top":"20px"},attrs:{type:"flex",justify:"end"}},[a("a-col",{attrs:{span:2}}),a("a-col",{attrs:{span:2}}),a("a-col",{attrs:{span:6}},[a("span",[e._v("平台：")]),a("a-radio-group",{staticStyle:{float:"right"},attrs:{value:e.os},on:{change:e.osChange}},[a("a-radio",{attrs:{value:"iOS"}},[e._v(" iOS ")]),a("a-radio",{attrs:{value:"Android"}},[e._v(" Android ")])],1)],1),a("a-col",{attrs:{span:4}},[a("a-button",{staticStyle:{float:"right","margin-top":"-5px"},attrs:{icon:"upload",type:"primary",disabled:e.uploadVersionFileBtn,loading:e.loading},on:{click:e.uploadFile}},[e._v("上传 ")])],1)],1)],1)],1),a("a-tab-pane",{key:"hot_update",attrs:{tab:"热更文件上传","force-render":""}},[a("a-space",{staticStyle:{width:"100%"},attrs:{direction:"vertical"}},[a("a-upload-dragger",{attrs:{name:"file",multiple:!1,"default-file-list":e.defaultVersionFileList,"before-upload":e.beforeUpload}},[a("p",{staticClass:"ant-upload-drag-icon"},[a("a-icon",{attrs:{type:"inbox"}})],1),a("p",{staticClass:"ant-upload-text"},[e._v(" 单击或拖动文件到此区域上传 ")]),a("p",{staticClass:"ant-upload-hint"},[e._v(" 仅支持单个文件，且文件名必须为Release.zip ")])]),a("a-row",{staticStyle:{"margin-top":"20px"},attrs:{type:"flex",justify:"end"}},[a("a-col",{attrs:{span:2}}),a("a-col",{attrs:{span:2}}),a("a-col",{attrs:{span:6}},[a("span",[e._v("平台：")]),a("a-radio-group",{staticStyle:{float:"right"},attrs:{value:e.os},on:{change:e.osChange}},[a("a-radio",{attrs:{value:"iOS"}},[e._v(" iOS ")]),a("a-radio",{attrs:{value:"Android"}},[e._v(" Android ")])],1)],1),a("a-col",{attrs:{span:4}},[a("a-button",{staticStyle:{float:"right","margin-top":"-5px"},attrs:{icon:"upload",type:"primary",disabled:e.uploadVersionFileBtn,loading:e.loading},on:{click:e.uploadFile}},[e._v("上传 ")])],1)],1)],1)],1)],1),a("a-modal",{attrs:{title:"修改版本信息",visible:e.visible,width:"50%"},on:{cancel:function(t){e.visible=!1}}},[a("template",{slot:"footer"},[a("a-button",{key:"back",on:{click:function(t){e.visible=!1}}},[e._v(" 取消 ")]),a("a-button",{key:"submit",attrs:{type:"primary",loading:e.loading},on:{click:e.save}},[e._v(" 保存 ")])],1),a("a-form-model",{ref:"ruleForm",attrs:{model:e.form,rules:e.rules,"label-col":e.labelCol,"wrapper-col":e.wrapperCol}},[a("a-form-model-item",{attrs:{label:"主版本号",prop:"version","has-feedback":!0}},[a("a-input",{model:{value:e.form.version,callback:function(t){e.$set(e.form,"version",t)},expression:"form.version"}})],1),e.form.audit?e._e():a("a-form-model-item",{attrs:{label:"最小版本号",prop:"min_version","has-feedback":!0}},[a("a-input",{model:{value:e.form.min_version,callback:function(t){e.$set(e.form,"min_version",t)},expression:"form.min_version"}})],1),a("a-form-model-item",{attrs:{label:"网关",prop:"gateway","has-feedback":!0}},[a("a-input",{model:{value:e.form.gateway,callback:function(t){e.$set(e.form,"gateway",t)},expression:"form.gateway"}})],1),a("a-form-model-item",{attrs:{label:"热更文件地址",prop:"cdn","has-feedback":!0}},[a("a-input",{model:{value:e.form.cdn,callback:function(t){e.$set(e.form,"cdn",t)},expression:"form.cdn"}})],1),a("a-form-model-item",{attrs:{label:"公告地址",prop:"announcement","has-feedback":!0}},[a("a-input",{model:{value:e.form.announcement,callback:function(t){e.$set(e.form,"announcement",t)},expression:"form.announcement"}})],1),a("a-form-model-item",{attrs:{label:"版本文件名",prop:"version_file","has-feedback":!0}},[a("a-input",{attrs:{placeholder:"请输入版本文件名称（带后缀）"},model:{value:e.form.version_file,callback:function(t){e.$set(e.form,"version_file",t)},expression:"form.version_file"}})],1)],1)],2)],1)},n=[],i=(a("ac1f"),a("1276"),a("96cf"),a("1da1")),o=a("7424"),s=a("b775");function c(e){return Object(s["e"])(o["VERSION_INFO"],s["a"].GET,{type:e})}function l(e){return Object(s["e"])(o["VERSION_SAVE"],s["a"].POST,e)}function u(e){return Object(s["e"])(o["VERSION_UPLOAD"],s["a"].POST,e)}function d(e){return Object(s["e"])(o["VERSION_UPLOAD_ZIP"],s["a"].POST,e)}var m={name:"Version",data:function(){return{labelCol:{span:5},wrapperCol:{span:13},loading:!1,visible:!1,form:{version:void 0,min_version:void 0,gateway:void 0,cdn:void 0,announcement:void 0,version_file:void 0,audit:!1,activate:!1},os:void 0,tabType:void 0,defaultVersionFileList:[],activate:!1,rules:{version:[{required:!0,message:"请输入客户端版本",trigger:"blur"},{min:1,max:10,message:"长度1~10",trigger:"blur"}],min_version:[{required:!0,message:"请输入客户端最小版本",trigger:"blur"},{min:1,max:10,message:"长度1~10",trigger:"blur"}],gateway:[{required:!0,message:"请输入网关地址",trigger:"blur"},{min:1,max:100,message:"长度1~100",trigger:"blur"}],cdn:[{required:!0,message:"请输入热更文件地址",trigger:"blur"},{min:1,max:100,message:"长度1~100",trigger:"blur"}],announcement:[{required:!0,message:"请输入公告地址",trigger:"blur"},{min:1,max:100,message:"长度1~100",trigger:"blur"}],version_file:[{required:!0,message:"请输入版本文件名",trigger:"blur"},{min:1,max:100,message:"长度1~100",trigger:"blur"}]}}},created:function(){this.fetch()},mounted:function(){},computed:{uploadVersionFileBtn:function(){return!this.os||!this.file}},methods:{fetch:function(){var e=this;return Object(i["a"])(regeneratorRuntime.mark((function t(){var a,r,n,i,o,s,l,u,d,m,f,p;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return t.prev=0,t.next=3,c(e.tabType);case 3:if(a=t.sent,r=a.data,n=r.code,i=r.msg,0===n){t.next=9;break}return e.$message.error(i),t.abrupt("return");case 9:o=r.data||{},s=o.version,l=o.min_version,u=o.gateway,d=o.cdn,m=o.announcement,f=o.version_file,p=o.activate,e.form.version=s,e.form.min_version=l,e.form.gateway=u,e.form.cdn=d,e.form.announcement=m,e.form.version_file=f,e.form.activate=p,e.activate=p,t.next=23;break;case 20:t.prev=20,t.t0=t["catch"](0),e.$message.error(t.t0.message);case 23:case"end":return t.stop()}}),t,null,[[0,20]])})))()},modify:function(e){this.visible=!0,this.form.audit=e},save:function(){var e=this;return Object(i["a"])(regeneratorRuntime.mark((function t(){var a,r,n,i;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:if(t.prev=0,e.checkMinVersion()){t.next=3;break}return t.abrupt("return");case 3:return e.loading=!0,t.next=6,l(e.form);case 6:if(a=t.sent,r=a.data,n=r.code,i=r.msg,0===n){t.next=13;break}return e.$message.error(i),e.loading=!1,t.abrupt("return");case 13:return e.$message.success("保存成功"),e.visible=!1,t.next=17,e.fetch();case 17:e.loading=!1,t.next=24;break;case 20:t.prev=20,t.t0=t["catch"](0),e.loading=!1,e.$message.error(t.t0.message);case 24:case"end":return t.stop()}}),t,null,[[0,20]])})))()},checkMinVersion:function(){if(this.form.audit)return!0;var e=this.form.version.split("."),t=this.form.min_version.split(".");if(3!==e.length)return this.$message.error("主版本号格式有误"),!1;if(3!==t.length)return this.$message.error("最小版本号格式有误"),!1;if(e[2]!==t[2])return this.$message.error("版本号最后一位必须相同"),!1;var a=1e4*e[0]+e[1],r=1e4*t[0]+t[1];return!(r>a)||(this.$message.error("最小版本号不能大于主版本号"),!1)},beforeUpload:function(e){return this.file=e,this.defaultVersionFileList=[e],!1},uploadFile:function(){var e=this;return Object(i["a"])(regeneratorRuntime.mark((function t(){var a,r,n,i,o,s,c;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:if(t.prev=0,a=e.file,r=new FormData,r.append("file",a),r.append("os",e.os),"version"!==e.tabType){t.next=9;break}n=u,t.next=15;break;case 9:if("hot_update"!==e.tabType){t.next=13;break}n=d,t.next=15;break;case 13:return e.$message.error("平台有误"),t.abrupt("return");case 15:return e.loading=!0,t.next=18,n(r);case 18:if(i=t.sent,o=i.data,s=o.code,c=o.msg,0===s){t.next=25;break}return e.loading=!1,e.$message.error(c),t.abrupt("return");case 25:e.$success({title:"提示",content:"上传成功"}),e.loading=!1,t.next=33;break;case 29:t.prev=29,t.t0=t["catch"](0),e.loading=!1,e.$message.error(t.t0.message);case 33:case"end":return t.stop()}}),t,null,[[0,29]])})))()},osChange:function(e){this.os=e.target.value},tabChange:function(e){var t=this;return Object(i["a"])(regeneratorRuntime.mark((function a(){return regeneratorRuntime.wrap((function(a){while(1)switch(a.prev=a.next){case 0:return t.tabType=e,t.form.audit="audit"===e,a.next=4,t.fetch();case 4:case"end":return a.stop()}}),a)})))()},switchChange:function(e){this.activate=e,this.form.activate=e,this.save()}}},f=m,p=(a("7adb"),a("0c7c")),v=Object(p["a"])(f,r,n,!1,null,"046ccba5",null);t["default"]=v.exports}}]);