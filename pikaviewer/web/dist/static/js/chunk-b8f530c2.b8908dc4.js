(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-b8f530c2"],{"01c3":function(t,e,a){"use strict";var n=a("6ce6"),o=a.n(n);o.a},"4ffd":function(t,e,a){t.exports=a.p+"static/img/logo.9ebb0df3.png"},5476:function(t,e,a){},"613e":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"footer"},[a("div",{staticClass:"links"},t._l(t.linkList,(function(e,n){return a("a",{key:n,attrs:{target:"_blank",href:e.link?e.link:"javascript: void(0)"}},[e.icon?a("a-icon",{attrs:{type:e.icon}}):t._e(),t._v(t._s(e.name)+" ")],1)})),0),a("div",{staticClass:"copyright"},[t._v(" Copyright"),a("a-icon",{attrs:{type:"copyright"}}),t._v(t._s(t.copyright)+" ")],1)])},o=[],r={name:"PageFooter",props:["copyright","linkList"]},c=r,i=(a("01c3"),a("0c7c")),s=Object(i["a"])(c,n,o,!1,null,"7b4663f1",null);e["a"]=s.exports},"6ce6":function(t,e,a){},"6d14":function(t,e,a){"use strict";var n=a("5476"),o=a.n(n);o.a},7424:function(t,e){var a="https://10.23.20.53:9991/v1",n="http://127.0.0.1:8888";t.exports={LOGIN:"".concat(a,"/admin/user/login"),ROUTES:"".concat(a,"/routes"),NODES_MODULE:"".concat(a,"/node/module"),NODES_PIKA:"".concat(a,"/node/pika"),QUERY_DATA:"".concat(a,"/query/data"),All_API:"".concat(n,"/all"),HISTORY:"".concat(n,"/history"),REQUEST_API:"".concat(n,"/"),PID:"".concat(a,"/restart/pid"),RESTART:"".concat(a,"/restart/do"),LOG:"".concat(a,"/restart/log"),PLAYER_INFO:"".concat(a,"/admin/player/info"),MAIL_INFO:"".concat(a,"/admin/player/mail/info"),MAIL_SEND:"".concat(a,"/admin/player/mail/send"),MAIL_DEL:"".concat(a,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(a,"/admin/player/mail/entire"),PLAYER_KICK_OFF:"".concat(a,"/admin/player/kick/off"),ANNOUNCEMENT_INFO:"".concat(a,"/announcement"),ANNOUNCEMENT_LIST:"".concat(a,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(a,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(a,"/announcement/del"),BATTLE_LOG_FILES:"".concat(a,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(a,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(a,"/battle/log/download"),BUILD_DO:"".concat(a,"/build/do"),BUILD_LOG:"".concat(a,"/build/log"),VERSION_INFO:"".concat(a,"/version/info"),VERSION_SAVE:"".concat(a,"/version/save"),VERSION_UPLOAD:"".concat(a,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(a,"/version/zip"),BROADCAST_SEND:"".concat(a,"/broadcast/send"),WL_LIST:"".concat(a,"/whitelist/list"),WL_SAVE:"".concat(a,"/whitelist/save"),WL_DEL:"".concat(a,"/whitelist/del"),GITLAB_MEMBERS:"".concat(a,"/gitlab/members"),GITLAB_MODIFY:"".concat(a,"/gitlab/modify"),CDKEY_SAVE:"".concat(a,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(a,"/cdkey/deactive"),BWL_LIST:"".concat(a,"/beta/whitelist/list"),BWL_SAVE:"".concat(a,"/beta/whitelist/save"),BWL_DEL:"".concat(a,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(a,"/map/line/info")}},"80c1":function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("common-layout",[n("div",{staticClass:"top"},[n("div",{staticClass:"header"},[n("img",{staticClass:"logo",attrs:{alt:"logo",src:a("4ffd")}}),n("span",{staticClass:"title"},[t._v(t._s(t.systemName))])])]),n("div",{staticClass:"login"},[n("a-form",{attrs:{form:t.form},on:{submit:t.onSubmit}},[n("a-tabs",{staticStyle:{padding:"0 2px"},attrs:{size:"large",tabBarStyle:{textAlign:"center"}}},[n("a-tab-pane",{key:"1",attrs:{tab:"账户密码登录"}},[n("a-alert",{directives:[{name:"show",rawName:"v-show",value:t.error,expression:"error"}],staticStyle:{"margin-bottom":"24px"},attrs:{type:"error",closable:!0,message:t.error,showIcon:""}}),n("a-form-item",{attrs:{"has-feedback":""}},[n("a-input",{directives:[{name:"decorator",rawName:"v-decorator",value:["name",{initialValue:"",rules:[{required:!0,message:"请输入账户名",whitespace:!0}]}],expression:"['name', {initialValue:'',rules: [{ required: true, message: '请输入账户名', whitespace: true}]}]"}],attrs:{autocomplete:"autocomplete",size:"large",placeholder:"请输入登录账号"}},[n("a-icon",{attrs:{slot:"prefix",type:"user"},slot:"prefix"})],1)],1),n("a-form-item",{attrs:{"has-feedback":""}},[n("a-input",{directives:[{name:"decorator",rawName:"v-decorator",value:["password",{initialValue:"",rules:[{required:!0,message:"请输入密码",whitespace:!0}]}],expression:"['password', {initialValue:'', rules: [{ required: true, message: '请输入密码', whitespace: true}]}]"}],attrs:{size:"large",placeholder:"请输入登录密码",autocomplete:"autocomplete",type:"password"}},[n("a-icon",{attrs:{slot:"prefix",type:"lock"},slot:"prefix"})],1)],1)],1)],1),n("div"),n("a-form-item",[n("a-button",{staticStyle:{width:"100%","margin-top":"24px"},attrs:{loading:t.logging,size:"large",htmlType:"submit",type:"primary"}},[t._v("登录 ")])],1)],1)],1)])},o=[],r=a("5530"),c=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"common-layout"},[a("div",{staticClass:"content"},[t._t("default")],2),a("page-footer",{attrs:{"link-list":t.footerLinks,copyright:t.copyright}})],1)},i=[],s=a("613e"),l=a("5880"),u={name:"CommonLayout",components:{PageFooter:s["a"]},computed:Object(r["a"])({},Object(l["mapState"])("setting",["footerLinks","copyright"]))},m=u,p=(a("6d14"),a("0c7c")),d=Object(p["a"])(m,c,i,!1,null,"1c265f57",null),f=d.exports,g=a("b13a"),b=a("b775"),h=a("89a5"),v={name:"Login",components:{CommonLayout:f},data:function(){return{logging:!1,error:"",form:this.$form.createForm(this)}},computed:{systemName:function(){return this.$store.state.setting.systemName}},mounted:function(){},methods:Object(r["a"])(Object(r["a"])({},Object(l["mapMutations"])("account",["setUser","setPermissions","setRoles"])),{},{onSubmit:function(t){var e=this;t.preventDefault(),this.form.validateFields((function(t){if(!t){e.logging=!0;var a=e.form.getFieldValue("name"),n=e.form.getFieldValue("password");Object(g["b"])(a,n).then(e.afterLogin).catch((function(t){e.logging=!1,e.$message.error("登录失败："+t.message)}))}}))},afterLogin:function(t){var e=this;this.logging=!1;var a=t.data,n=t.data,o=n.code,r=n.msg;if(0!==o)return this.$message.error(r);if(a.code>=0){var c=a.data;c.avatar="https://gw.alipayobjects.com/zos/rmsportal/BiazfanxmamNRoxxVxka.png",this.setUser(c),Object(b["f"])({token:a.data.token,expireAt:new Date(a.data.expireAt)}),Object(g["a"])().then((function(t){var a=t.data,n=a.code,o=a.msg,r=a.data;if(0!==n)return e.$message.error(o);Object(h["d"])(r),2===c.role?e.$router.push("/mail/query"):e.$router.push("/restart"),e.$message.success("登录成功",3)}))}else this.error=a.msg}})},_=v,E=(a("93b6"),Object(p["a"])(_,n,o,!1,null,"c6abea9e",null)),O=E.exports;e["default"]=O},"93b6":function(t,e,a){"use strict";var n=a("995b"),o=a.n(n);o.a},"995b":function(t,e,a){},b13a:function(t,e,a){"use strict";a.d(e,"b",(function(){return c})),a.d(e,"a",(function(){return s})),a.d(e,"c",(function(){return u}));a("96cf");var n=a("1da1"),o=a("7424"),r=a("b775");function c(t,e){return i.apply(this,arguments)}function i(){return i=Object(n["a"])(regeneratorRuntime.mark((function t(e,a){return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return t.abrupt("return",Object(r["e"])(o["LOGIN"],r["a"].POST,{username:e,password:a}));case 1:case"end":return t.stop()}}),t)}))),i.apply(this,arguments)}function s(){return l.apply(this,arguments)}function l(){return l=Object(n["a"])(regeneratorRuntime.mark((function t(){return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return t.abrupt("return",Object(r["e"])(o["ROUTES"],r["a"].GET));case 1:case"end":return t.stop()}}),t)}))),l.apply(this,arguments)}function u(){localStorage.removeItem("admin.routes"),localStorage.removeItem("admin.permissions"),localStorage.removeItem("admin.roles"),Object(r["d"])()}}}]);