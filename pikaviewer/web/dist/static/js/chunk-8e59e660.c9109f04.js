(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-8e59e660"],{"24d73":function(e,t,r){"use strict";r.d(t,"d",(function(){return i})),r.d(t,"c",(function(){return o})),r.d(t,"e",(function(){return s})),r.d(t,"a",(function(){return c})),r.d(t,"b",(function(){return d}));var a=r("7424"),n=r("b775");function i(e){return Object(n["e"])(a["PLAYER_INFO"],n["a"].GET,{name:e})}function o(e){return Object(n["e"])(a["MAIL_INFO"],n["a"].GET,{uid:e})}function s(e){return Object(n["e"])(a["MAIL_SEND"],n["a"].POST,e)}function c(e){return Object(n["e"])(a["MAIL_DEL"],n["a"].POST,e)}function d(){return Object(n["e"])(a["MAIL_ENTIRE"],n["a"].GET)}},"498a":function(e,t,r){"use strict";var a=r("23e7"),n=r("58a8").trim,i=r("c8d2");a({target:"String",proto:!0,forced:i("trim")},{trim:function(){return n(this)}})},5899:function(e,t){e.exports="\t\n\v\f\r                　\u2028\u2029\ufeff"},"58a8":function(e,t,r){var a=r("1d80"),n=r("5899"),i="["+n+"]",o=RegExp("^"+i+i+"*"),s=RegExp(i+i+"*$"),c=function(e){return function(t){var r=String(a(t));return 1&e&&(r=r.replace(o,"")),2&e&&(r=r.replace(s,"")),r}};e.exports={start:c(1),end:c(2),trim:c(3)}},7424:function(e,t){var r="https://10.23.20.53:9991/v1",a="http://127.0.0.1:8888";e.exports={LOGIN:"".concat(r,"/admin/user/login"),ROUTES:"".concat(r,"/routes"),NODES_MODULE:"".concat(r,"/node/module"),NODES_PIKA:"".concat(r,"/node/pika"),QUERY_DATA:"".concat(r,"/query/data"),All_API:"".concat(a,"/all"),HISTORY:"".concat(a,"/history"),REQUEST_API:"".concat(a,"/"),PID:"".concat(r,"/restart/pid"),RESTART:"".concat(r,"/restart/do"),LOG:"".concat(r,"/restart/log"),PLAYER_INFO:"".concat(r,"/admin/player/info"),MAIL_INFO:"".concat(r,"/admin/player/mail/info"),MAIL_SEND:"".concat(r,"/admin/player/mail/send"),MAIL_DEL:"".concat(r,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(r,"/admin/player/mail/entire"),ANNOUNCEMENT_INFO:"".concat(r,"/announcement"),ANNOUNCEMENT_LIST:"".concat(r,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(r,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(r,"/announcement/del"),BATTLE_LOG_FILES:"".concat(r,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(r,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(r,"/battle/log/download"),BUILD_DO:"".concat(r,"/build/do"),BUILD_LOG:"".concat(r,"/build/log"),VERSION_INFO:"".concat(r,"/version/info"),VERSION_SAVE:"".concat(r,"/version/save"),VERSION_UPLOAD:"".concat(r,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(r,"/version/zip"),BROADCAST_SEND:"".concat(r,"/broadcast/send"),WL_LIST:"".concat(r,"/whitelist/list"),WL_SAVE:"".concat(r,"/whitelist/save"),WL_DEL:"".concat(r,"/whitelist/del"),GITLAB_MEMBERS:"".concat(r,"/gitlab/members"),GITLAB_MODIFY:"".concat(r,"/gitlab/modify"),CDKEY_SAVE:"".concat(r,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(r,"/cdkey/deactive"),BWL_LIST:"".concat(r,"/beta/whitelist/list"),BWL_SAVE:"".concat(r,"/beta/whitelist/save"),BWL_DEL:"".concat(r,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(r,"/map/line/info")}},a9e3:function(e,t,r){"use strict";var a=r("83ab"),n=r("da84"),i=r("94ca"),o=r("6eeb"),s=r("5135"),c=r("c6b6"),d=r("7156"),m=r("c04e"),l=r("d039"),f=r("7c73"),u=r("241c").f,h=r("06cf").f,p=r("9bf2").f,v=r("58a8").trim,g="Number",b=n[g],I=b.prototype,A=c(f(I))==g,E=function(e){var t,r,a,n,i,o,s,c,d=m(e,!1);if("string"==typeof d&&d.length>2)if(d=v(d),t=d.charCodeAt(0),43===t||45===t){if(r=d.charCodeAt(2),88===r||120===r)return NaN}else if(48===t){switch(d.charCodeAt(1)){case 66:case 98:a=2,n=49;break;case 79:case 111:a=8,n=55;break;default:return+d}for(i=d.slice(2),o=i.length,s=0;s<o;s++)if(c=i.charCodeAt(s),c<48||c>n)return NaN;return parseInt(i,a)}return+d};if(i(g,!b(" 0o1")||!b("0b1")||b("+0x1"))){for(var N,O=function(e){var t=arguments.length<1?0:e,r=this;return r instanceof O&&(A?l((function(){I.valueOf.call(r)})):c(r)!=g)?d(new b(E(t)),r,O):E(t)},_=a?u(b):"MAX_VALUE,MIN_VALUE,NaN,NEGATIVE_INFINITY,POSITIVE_INFINITY,EPSILON,isFinite,isInteger,isNaN,isSafeInteger,MAX_SAFE_INTEGER,MIN_SAFE_INTEGER,parseFloat,parseInt,isInteger".split(","),D=0;_.length>D;D++)s(b,N=_[D])&&!s(O,N)&&p(O,N,h(b,N));O.prototype=I,I.constructor=O,o(n,g,O)}},bec6:function(e,t,r){"use strict";r.r(t);var a=function(){var e=this,t=e.$createElement,r=e._self._c||t;return r("a-card",[r("a-form-model",{ref:"form",attrs:{model:e.form,rules:e.rules,"label-col":e.labelCol,"wrapper-col":e.wrapperCol}},[r("a-form-model-item",{attrs:{label:"发送方式"}},[r("a-radio-group",{attrs:{"default-value":e.form.sendWay},on:{change:e.sendWayChange}},[r("a-radio",{attrs:{value:1}},[e._v(" 发给指定玩家 ")]),r("a-radio",{attrs:{value:2}},[e._v(" 全服邮件 ")])],1)],1),1===e.form.sendWay?r("a-form-model-item",{attrs:{label:"UID",prop:"uid"}},[r("a-input",{attrs:{type:"textarea",placeholder:"输入玩家id，多个玩家id用空格分割"},model:{value:e.form.uid,callback:function(t){e.$set(e.form,"uid",t)},expression:"form.uid"}})],1):e._e(),r("a-form-model-item",{attrs:{label:"发送者"}},[r("a-input",{attrs:{placeholder:"用于客户端显示是谁发送的这封邮件（选填）"},model:{value:e.form.sender,callback:function(t){e.$set(e.form,"sender",t)},expression:"form.sender"}})],1),r("a-form-model-item",{attrs:{label:"标题",prop:"title"}},[r("a-input",{attrs:{placeholder:"请输入邮件标题"},model:{value:e.form.title,callback:function(t){e.$set(e.form,"title",t)},expression:"form.title"}})],1),r("a-form-model-item",{attrs:{label:"文本ID"}},[r("a-input",{attrs:{placeholder:"文本id和邮件内容2选1"},model:{value:e.form.textId,callback:function(t){e.$set(e.form,"textId",t)},expression:"form.textId"}})],1),r("a-form-model-item",{attrs:{label:"邮件内容"}},[r("a-input",{attrs:{type:"textarea",placeholder:"文本id和邮件内容2选1"},model:{value:e.form.content,callback:function(t){e.$set(e.form,"content",t)},expression:"form.content"}})],1),r("a-form-model-item",{attrs:{label:"有效时间",required:""}},[r("a-date-picker",{attrs:{"show-time":"",format:"YYYY-MM-DD HH:mm:ss",placeholder:"请选择邮件激活时间"},on:{openChange:e.handleActivatedAtOpenChange},model:{value:e.form.activatedAt,callback:function(t){e.$set(e.form,"activatedAt",t)},expression:"form.activatedAt"}}),e._v(" ~ "),r("a-date-picker",{attrs:{"show-time":"",format:"YYYY-MM-DD HH:mm:ss",placeholder:"请选择邮件过期时间",open:e.expiredAtOpen},on:{openChange:e.handleExpiredAtOpenOpenChange},model:{value:e.form.expiredAt,callback:function(t){e.$set(e.form,"expiredAt",t)},expression:"form.expiredAt"}})],1),r("a-form-model-item",{attrs:{label:"金币"}},[r("a-input-number",{staticStyle:{width:"100%"},attrs:{min:0},model:{value:e.form.gold,callback:function(t){e.$set(e.form,"gold",t)},expression:"form.gold"}})],1),r("a-form-model-item",{attrs:{label:"绑钻"}},[r("a-input-number",{staticStyle:{width:"100%"},attrs:{min:0},model:{value:e.form.bindDiamond,callback:function(t){e.$set(e.form,"bindDiamond",t)},expression:"form.bindDiamond"}})],1),r("a-form-model-item",{attrs:{label:"非绑钻"}},[r("a-input-number",{staticStyle:{width:"100%"},attrs:{min:0},model:{value:e.form.diamond,callback:function(t){e.$set(e.form,"diamond",t)},expression:"form.diamond"}})],1),r("a-form-model-item",{attrs:{label:"额外附件"}},[r("a-input",{attrs:{type:"textarea",placeholder:"物品id和物品数量逗号分割，多个物品用分号分割。例如：1001,1;1002,2 （选填）"},model:{value:e.form.attachment,callback:function(t){e.$set(e.form,"attachment",t)},expression:"form.attachment"}})],1),r("a-form-model-item",{attrs:{"wrapper-col":{span:14,offset:4}}},[r("a-button",{attrs:{type:"primary",loading:e.loading},on:{click:e.onSubmit}},[e._v(" 发送 ")]),r("a-button",{staticStyle:{"margin-left":"10px"},on:{click:e.resetForm}},[e._v(" 重置 ")])],1)],1)],1)},n=[],i=(r("c975"),r("a15b"),r("a9e3"),r("ac1f"),r("1276"),r("498a"),r("96cf"),r("1da1")),o=r("b85c"),s=r("24d73"),c={name:"MailSend",components:{},data:function(){return{labelCol:{span:4},wrapperCol:{span:14},loading:!1,expiredAtOpen:!1,form:{sendWay:1,uid:void 0,sender:void 0,textId:void 0,title:void 0,content:void 0,activatedAt:void 0,expiredAt:void 0,gold:0,bindDiamond:0,diamond:0,attachment:void 0},rules:{uid:[{required:!0,message:"请输入接收邮件的玩家id",trigger:"blur"}],title:[{required:!0,message:"请输入邮件标题",trigger:"blur"}],activatedAt:[{required:!0,message:"请选择时间",trigger:"change"}]}}},methods:{sendWayChange:function(e){this.form.sendWay=e.target.value},formatData:function(){var e={sendWay:this.form.sendWay,sender:(this.form.sender||"").trim(),title:(this.form.title||"").trim(),content:(this.form.content||"").trim(),hi:(this.form.hi||"").trim()};if(1!==this.form.sendWay||this.form.uid&&this.form.uid.trim())if(this.form.title&&this.form.title.trim()){var t,r=(this.form.uid||"").trim().split(" "),a=[],n={},i=Object(o["a"])(r);try{for(i.s();!(t=i.n()).done;){var s=t.value,c=Number(s.trim());if(isNaN(c))return void this.$message.error("存在无效的UID："+s);if(n[c])return void this.$message.error("存在重复的UID："+c);n[c]=!0,a.push(c)}}catch(_){i.e(_)}finally{i.f()}if(a.length<=0)this.$message.error("未解析到有效的UID");else if(a.length>100)this.$message.error("一次最多支持100个玩家发放");else{if(e.uid=a,this.form.textId){var d=+this.form.textId.trim();if(isNaN(d))return void this.$message.error("文本ID只能是数字");if(d<=0)return void this.$message.error("文本ID只能是正整数");if(String(d).indexOf(".")>-1)return void this.$message.error("文本ID只能是正整数");e.textId=d}if(e.content=(this.form.content||"").trim(),e.textId||e.content)if(e.textId&&e.content)this.$message.error("文本ID和邮件内容只能包含一个");else{var m=(this.form.attachment||"").trim();if(m){var l,f={},u=!1,h=m.split(";"),p=Object(o["a"])(h);try{for(p.s();!(l=p.n()).done;){var v=l.value;if(v=v.trim().split(","),2!==v.length)return void this.$message.error("附件数据格式有误");if(v[0].indexOf(".")>-1)return void this.$message.error("物品ID只能是正整数");if(v[1].indexOf(".")>-1)return void this.$message.error("物品数量只能是正整数");var g=+v[0].trim(),b=+v[1].trim();if(isNaN(g)||g<=0)return void this.$message.error("物品ID只能是正整数");if(isNaN(b)||b<=0)return void this.$message.error("物品数量只能是正整数");f[g]=b,u=!0}}catch(_){p.e(_)}finally{p.f()}u&&(e.attachment=f)}if(e.attachment||(e.attachment={}),this.form.gold>0&&(e.attachment[101]=this.form.gold),this.form.bindDiamond>0&&(e.attachment[102]=this.form.bindDiamond),this.form.diamond>0&&(e.attachment[103]=this.form.diamond),this.form.args&&this.form.args.trim()){var I,A=this.form.args.trim().split(" "),E=[],N=Object(o["a"])(A);try{for(N.s();!(I=N.n()).done;){var O=I.value;E.push(O.trim())}}catch(_){N.e(_)}finally{N.f()}e.args=E}if(this.form.activatedAt)if(this.form.expiredAt)if(new Date(this.form.activatedAt).getTime()>=new Date(this.form.expiredAt).getTime())this.$message.error("激活时间不能大于过期时间");else{if(!(new Date(this.form.expiredAt).getTime()<=Date.now()))return e;this.$message.error("过期时间不能小于当前时间")}else this.$message.error("请选择邮件的过期时间");else this.$message.error("请选择邮件的激活时间")}else this.$message.error("文本ID和邮件内容必须包含一个")}}else this.$message.error("请输入邮件标题");else this.$message.error("请输入接收邮件的玩家id")},onSubmit:function(){var e=this;return Object(i["a"])(regeneratorRuntime.mark((function t(){var r,a,n,i,o,c;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:if(r=e.formatData(),r){t.next=3;break}return t.abrupt("return");case 3:return e.loading=!0,t.prev=4,r.activatedAt=new Date(e.form.activatedAt).getTime(),r.expiredAt=new Date(e.form.expiredAt).getTime(),t.next=9,Object(s["e"])(r);case 9:if(a=t.sent,n=a.data,i=n.code,o=n.msg,0===i){t.next=16;break}return e.loading=!1,e.$message.error(o),t.abrupt("return");case 16:c=n.data||[],e.loading=!1,c.length<=0?e.$message.success("发送成功"):e.$confirm({title:"如下UID玩家发送失败",content:c.join(","),okType:"danger"}),t.next=25;break;case 21:t.prev=21,t.t0=t["catch"](4),e.loading=!1,e.$message.error(t.t0.message);case 25:case"end":return t.stop()}}),t,null,[[4,21]])})))()},resetForm:function(){this.$refs.form.resetFields(),this.form={sendWay:1,uid:void 0,sender:void 0,textId:void 0,title:void 0,content:void 0,activatedAt:void 0,expiredAt:void 0,gold:0,bindDiamond:0,diamond:0,attachment:void 0}},handleActivatedAtOpenChange:function(e){e||(this.expiredAtOpen=!0)},handleExpiredAtOpenOpenChange:function(e){this.expiredAtOpen=e}}},d=c,m=r("0c7c"),l=Object(m["a"])(d,a,n,!1,null,"a7819e14",null);t["default"]=l.exports},c8d2:function(e,t,r){var a=r("d039"),n=r("5899"),i="​᠎";e.exports=function(e){return a((function(){return!!n[e]()||i[e]()!=i||n[e].name!==e}))}}}]);