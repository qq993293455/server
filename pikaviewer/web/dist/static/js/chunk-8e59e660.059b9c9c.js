(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-8e59e660"],{"24d73":function(t,e,r){"use strict";r.d(e,"e",(function(){return i})),r.d(e,"d",(function(){return o})),r.d(e,"f",(function(){return s})),r.d(e,"a",(function(){return c})),r.d(e,"b",(function(){return d})),r.d(e,"c",(function(){return l}));var a=r("7424"),n=r("b775");function i(t){return Object(n["e"])(a["PLAYER_INFO"],n["a"].GET,{name:t})}function o(t){return Object(n["e"])(a["MAIL_INFO"],n["a"].GET,{uid:t})}function s(t){return Object(n["e"])(a["MAIL_SEND"],n["a"].POST,t)}function c(t){return Object(n["e"])(a["MAIL_DEL"],n["a"].POST,t)}function d(){return Object(n["e"])(a["MAIL_ENTIRE"],n["a"].GET)}function l(t,e){return Object(n["e"])(a["PLAYER_KICK_OFF"],n["a"].GET,{uid:t,sec:e})}},"498a":function(t,e,r){"use strict";var a=r("23e7"),n=r("58a8").trim,i=r("c8d2");a({target:"String",proto:!0,forced:i("trim")},{trim:function(){return n(this)}})},5899:function(t,e){t.exports="\t\n\v\f\r                　\u2028\u2029\ufeff"},"58a8":function(t,e,r){var a=r("1d80"),n=r("5899"),i="["+n+"]",o=RegExp("^"+i+i+"*"),s=RegExp(i+i+"*$"),c=function(t){return function(e){var r=String(a(e));return 1&t&&(r=r.replace(o,"")),2&t&&(r=r.replace(s,"")),r}};t.exports={start:c(1),end:c(2),trim:c(3)}},7424:function(t,e){var r="v1",a="http://127.0.0.1:8888";t.exports={SYSTEM_NAME:"".concat(r,"/name"),LOGIN:"".concat(r,"/admin/user/login"),ROUTES:"".concat(r,"/routes"),NODES_MODULE:"".concat(r,"/node/module"),NODES_PIKA:"".concat(r,"/node/pika"),QUERY_DATA:"".concat(r,"/query/data"),All_API:"".concat(a,"/all"),HISTORY:"".concat(a,"/history"),REQUEST_API:"".concat(a,"/"),PID:"".concat(r,"/restart/pid"),RESTART:"".concat(r,"/restart/do"),LOG:"".concat(r,"/restart/log"),PLAYER_INFO:"".concat(r,"/admin/player/info"),MAIL_INFO:"".concat(r,"/admin/player/mail/info"),MAIL_SEND:"".concat(r,"/admin/player/mail/send"),MAIL_DEL:"".concat(r,"/admin/player/mail/delete"),MAIL_ENTIRE:"".concat(r,"/admin/player/mail/entire"),PLAYER_KICK_OFF:"".concat(r,"/admin/player/kick/off"),ANNOUNCEMENT_INFO:"".concat(r,"/announcement"),ANNOUNCEMENT_LIST:"".concat(r,"/announcement/list"),ANNOUNCEMENT_SAVE:"".concat(r,"/announcement/save"),ANNOUNCEMENT_DEL:"".concat(r,"/announcement/del"),BATTLE_LOG_FILES:"".concat(r,"/battle/log/files"),BATTLE_LOG_DEL:"".concat(r,"/battle/log/del"),BATTLE_LOG_DOWNLOAD:"".concat(r,"/battle/log/download"),BUILD_DO:"".concat(r,"/build/do"),BUILD_LOG:"".concat(r,"/build/log"),VERSION_INFO:"".concat(r,"/version/info"),VERSION_SAVE:"".concat(r,"/version/save"),VERSION_UPLOAD:"".concat(r,"/version/file"),VERSION_UPLOAD_ZIP:"".concat(r,"/version/zip"),BROADCAST_SEND:"".concat(r,"/broadcast/send"),WL_LIST:"".concat(r,"/whitelist/list"),WL_SAVE:"".concat(r,"/whitelist/save"),WL_DEL:"".concat(r,"/whitelist/del"),GITLAB_MEMBERS:"".concat(r,"/gitlab/members"),GITLAB_MODIFY:"".concat(r,"/gitlab/modify"),CDKEY_SAVE:"".concat(r,"/cdkey/save"),CDKEY_DEACTIVE:"".concat(r,"/cdkey/deactive"),BWL_LIST:"".concat(r,"/beta/whitelist/list"),BWL_SAVE:"".concat(r,"/beta/whitelist/save"),BWL_DEL:"".concat(r,"/beta/whitelist/del"),MAP_LINE_INFO:"".concat(r,"/map/line/info")}},a9e3:function(t,e,r){"use strict";var a=r("83ab"),n=r("da84"),i=r("94ca"),o=r("6eeb"),s=r("5135"),c=r("c6b6"),d=r("7156"),l=r("c04e"),m=r("d039"),f=r("7c73"),u=r("241c").f,p=r("06cf").f,h=r("9bf2").f,v=r("58a8").trim,g="Number",b=n[g],_=b.prototype,A=c(f(_))==g,I=function(t){var e,r,a,n,i,o,s,c,d=l(t,!1);if("string"==typeof d&&d.length>2)if(d=v(d),e=d.charCodeAt(0),43===e||45===e){if(r=d.charCodeAt(2),88===r||120===r)return NaN}else if(48===e){switch(d.charCodeAt(1)){case 66:case 98:a=2,n=49;break;case 79:case 111:a=8,n=55;break;default:return+d}for(i=d.slice(2),o=i.length,s=0;s<o;s++)if(c=i.charCodeAt(s),c<48||c>n)return NaN;return parseInt(i,a)}return+d};if(i(g,!b(" 0o1")||!b("0b1")||b("+0x1"))){for(var E,N=function(t){var e=arguments.length<1?0:t,r=this;return r instanceof N&&(A?m((function(){_.valueOf.call(r)})):c(r)!=g)?d(new b(I(e)),r,N):I(e)},O=a?u(b):"MAX_VALUE,MIN_VALUE,NaN,NEGATIVE_INFINITY,POSITIVE_INFINITY,EPSILON,isFinite,isInteger,isNaN,isSafeInteger,MAX_SAFE_INTEGER,MIN_SAFE_INTEGER,parseFloat,parseInt,isInteger".split(","),x=0;O.length>x;x++)s(b,E=O[x])&&!s(N,E)&&h(N,E,p(b,E));N.prototype=_,_.constructor=N,o(n,g,N)}},bec6:function(t,e,r){"use strict";r.r(e);var a=function(){var t=this,e=t.$createElement,r=t._self._c||e;return r("a-card",[r("a-form-model",{ref:"form",attrs:{model:t.form,rules:t.rules,"label-col":t.labelCol,"wrapper-col":t.wrapperCol}},[r("a-form-model-item",{attrs:{label:"发送方式"}},[r("a-radio-group",{attrs:{"default-value":t.form.sendWay},on:{change:t.sendWayChange}},[r("a-radio",{attrs:{value:1}},[t._v(" 发给指定玩家 ")]),r("a-radio",{attrs:{value:2}},[t._v(" 全服邮件 ")])],1)],1),1===t.form.sendWay?r("a-form-model-item",{attrs:{label:"UID",prop:"uid"}},[r("a-input",{attrs:{type:"textarea",placeholder:"输入玩家id，多个玩家id用空格分割"},model:{value:t.form.uid,callback:function(e){t.$set(t.form,"uid",e)},expression:"form.uid"}})],1):t._e(),r("a-form-model-item",{attrs:{label:"发送者"}},[r("a-input",{attrs:{placeholder:"用于客户端显示是谁发送的这封邮件（选填）"},model:{value:t.form.sender,callback:function(e){t.$set(t.form,"sender",e)},expression:"form.sender"}})],1),r("a-form-model-item",{attrs:{label:"标题",prop:"title"}},[r("a-input",{attrs:{placeholder:"请输入邮件标题"},model:{value:t.form.title,callback:function(e){t.$set(t.form,"title",e)},expression:"form.title"}})],1),r("a-form-model-item",{attrs:{label:"文本ID"}},[r("a-input",{attrs:{placeholder:"文本id和邮件内容2选1"},model:{value:t.form.textId,callback:function(e){t.$set(t.form,"textId",e)},expression:"form.textId"}})],1),r("a-form-model-item",{attrs:{label:"英文内容"}},[r("a-input",{attrs:{type:"textarea",placeholder:"文本id和邮件内容2选1"},model:{value:t.form.content_en,callback:function(e){t.$set(t.form,"content_en",e)},expression:"form.content_en"}})],1),r("a-form-model-item",{attrs:{label:"中文内容"}},[r("a-input",{attrs:{type:"textarea",placeholder:"文本id和邮件内容2选1"},model:{value:t.form.content_cn,callback:function(e){t.$set(t.form,"content_cn",e)},expression:"form.content_cn"}})],1),r("a-form-model-item",{attrs:{label:"繁体内容"}},[r("a-input",{attrs:{type:"textarea",placeholder:"文本id和邮件内容2选1"},model:{value:t.form.content_traditional,callback:function(e){t.$set(t.form,"content_traditional",e)},expression:"form.content_traditional"}})],1),r("a-form-model-item",{attrs:{label:"有效时间",required:""}},[r("a-date-picker",{attrs:{"show-time":"",format:"YYYY-MM-DD HH:mm:ss",placeholder:"请选择邮件激活时间"},on:{openChange:t.handleActivatedAtOpenChange},model:{value:t.form.activatedAt,callback:function(e){t.$set(t.form,"activatedAt",e)},expression:"form.activatedAt"}}),t._v(" ~ "),r("a-date-picker",{attrs:{"show-time":"",format:"YYYY-MM-DD HH:mm:ss",placeholder:"请选择邮件过期时间",open:t.expiredAtOpen},on:{openChange:t.handleExpiredAtOpenOpenChange},model:{value:t.form.expiredAt,callback:function(e){t.$set(t.form,"expiredAt",e)},expression:"form.expiredAt"}})],1),r("a-form-model-item",{attrs:{label:"金币"}},[r("a-input-number",{staticStyle:{width:"100%"},attrs:{min:0},model:{value:t.form.gold,callback:function(e){t.$set(t.form,"gold",e)},expression:"form.gold"}})],1),r("a-form-model-item",{attrs:{label:"绑钻"}},[r("a-input-number",{staticStyle:{width:"100%"},attrs:{min:0},model:{value:t.form.bindDiamond,callback:function(e){t.$set(t.form,"bindDiamond",e)},expression:"form.bindDiamond"}})],1),r("a-form-model-item",{attrs:{label:"非绑钻"}},[r("a-input-number",{staticStyle:{width:"100%"},attrs:{min:0},model:{value:t.form.diamond,callback:function(e){t.$set(t.form,"diamond",e)},expression:"form.diamond"}})],1),r("a-form-model-item",{attrs:{label:"额外附件"}},[r("a-input",{attrs:{type:"textarea",placeholder:"物品id和物品数量逗号分割，多个物品用分号分割。例如：1001,1;1002,2 （选填）"},model:{value:t.form.attachment,callback:function(e){t.$set(t.form,"attachment",e)},expression:"form.attachment"}})],1),r("a-form-model-item",{attrs:{"wrapper-col":{span:14,offset:4}}},[r("a-button",{attrs:{type:"primary",loading:t.loading},on:{click:t.onSubmit}},[t._v(" 发送 ")]),r("a-button",{staticStyle:{"margin-left":"10px"},on:{click:t.resetForm}},[t._v(" 重置 ")])],1)],1)],1)},n=[],i=(r("c975"),r("a15b"),r("a9e3"),r("ac1f"),r("1276"),r("498a"),r("96cf"),r("1da1")),o=r("b85c"),s=r("24d73"),c={name:"MailSend",components:{},data:function(){return{labelCol:{span:4},wrapperCol:{span:14},loading:!1,expiredAtOpen:!1,form:{sendWay:1,uid:void 0,sender:void 0,textId:void 0,title:void 0,content:void 0,content_en:void 0,content_cn:void 0,content_traditional:void 0,activatedAt:void 0,expiredAt:void 0,gold:0,bindDiamond:0,diamond:0,attachment:void 0},rules:{uid:[{required:!0,message:"请输入接收邮件的玩家id",trigger:"blur"}],title:[{required:!0,message:"请输入邮件标题",trigger:"blur"}],activatedAt:[{required:!0,message:"请选择时间",trigger:"change"}]}}},methods:{sendWayChange:function(t){this.form.sendWay=t.target.value},formatData:function(){var t={send_way:this.form.sendWay,sender:(this.form.sender||"").trim(),title:(this.form.title||"").trim(),content:(this.form.content||"").trim(),hi:(this.form.hi||"").trim()};if(1!==this.form.sendWay||this.form.uid&&this.form.uid.trim())if(this.form.title&&this.form.title.trim()){var e,r=(this.form.uid||"").trim().split(" "),a=[],n={},i=Object(o["a"])(r);try{for(i.s();!(e=i.n()).done;){var s=e.value,c=Number(s.trim());if(isNaN(c))return void this.$message.error("存在无效的UID："+s);if(n[c])return void this.$message.error("存在重复的UID："+c);n[c]=!0,a.push(c)}}catch(L){i.e(L)}finally{i.f()}if(a.length<=0)this.$message.error("未解析到有效的UID");else if(a.length>100)this.$message.error("一次最多支持100个玩家发放");else{if(t.uid=a,this.form.textId){var d=+this.form.textId.trim();if(isNaN(d))return void this.$message.error("文本ID只能是数字");if(d<=0)return void this.$message.error("文本ID只能是正整数");if(String(d).indexOf(".")>-1)return void this.$message.error("文本ID只能是正整数");t.text_id=d}var l=(this.form.content_en||"").trim(),m=(this.form.content_cn||"").trim(),f=(this.form.content_traditional||"").trim();if(l&&m&&f&&(t.content=JSON.stringify({10:l,40:m,41:f})),t.text_id||t.content)if(t.text_id&&t.content)this.$message.error("文本ID和邮件内容只能包含一个");else{var u=(this.form.attachment||"").trim();if(u){var p,h={},v=!1,g=u.split(";"),b=Object(o["a"])(g);try{for(b.s();!(p=b.n()).done;){var _=p.value;if(_=_.trim().split(","),2!==_.length)return void this.$message.error("附件数据格式有误");if(_[0].indexOf(".")>-1)return void this.$message.error("物品ID只能是正整数");if(_[1].indexOf(".")>-1)return void this.$message.error("物品数量只能是正整数");var A=+_[0].trim(),I=+_[1].trim();if(isNaN(A)||A<=0)return void this.$message.error("物品ID只能是正整数");if(isNaN(I)||I<=0)return void this.$message.error("物品数量只能是正整数");h[A]=I,v=!0}}catch(L){b.e(L)}finally{b.f()}v&&(t.attachment=h)}if(t.attachment||(t.attachment={}),this.form.gold>0&&(t.attachment[101]=this.form.gold),this.form.bindDiamond>0&&(t.attachment[102]=this.form.bindDiamond),this.form.diamond>0&&(t.attachment[103]=this.form.diamond),this.form.args&&this.form.args.trim()){var E,N=this.form.args.trim().split(" "),O=[],x=Object(o["a"])(N);try{for(x.s();!(E=x.n()).done;){var D=E.value;O.push(D.trim())}}catch(L){x.e(L)}finally{x.f()}t.args=O}if(this.form.activatedAt)if(this.form.expiredAt)if(new Date(this.form.activatedAt).getTime()>=new Date(this.form.expiredAt).getTime())this.$message.error("激活时间不能大于过期时间");else{if(!(new Date(this.form.expiredAt).getTime()<=Date.now()))return t;this.$message.error("过期时间不能小于当前时间")}else this.$message.error("请选择邮件的过期时间");else this.$message.error("请选择邮件的激活时间")}else this.$message.error("文本ID和邮件内容必须包含一个")}}else this.$message.error("请输入邮件标题");else this.$message.error("请输入接收邮件的玩家id")},onSubmit:function(){var t=this;return Object(i["a"])(regeneratorRuntime.mark((function e(){var r,a,n,i,o,c;return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:if(r=t.formatData(),r){e.next=3;break}return e.abrupt("return");case 3:return t.loading=!0,e.prev=4,r.activated_at=new Date(t.form.activatedAt).getTime(),r.expired_at=new Date(t.form.expiredAt).getTime(),e.next=9,Object(s["f"])(r);case 9:if(a=e.sent,n=a.data,i=n.code,o=n.msg,0===i){e.next=16;break}return t.loading=!1,t.$message.error(o),e.abrupt("return");case 16:c=n.data||[],t.loading=!1,c.length<=0?t.$message.success("发送成功"):t.$confirm({title:"如下UID玩家发送失败",content:c.join(","),okType:"danger"}),e.next=25;break;case 21:e.prev=21,e.t0=e["catch"](4),t.loading=!1,t.$message.error(e.t0.message);case 25:case"end":return e.stop()}}),e,null,[[4,21]])})))()},resetForm:function(){this.$refs.form.resetFields(),this.form={sendWay:1,uid:void 0,sender:void 0,textId:void 0,title:void 0,content:void 0,activatedAt:void 0,expiredAt:void 0,gold:0,bindDiamond:0,diamond:0,attachment:void 0}},handleActivatedAtOpenChange:function(t){t||(this.expiredAtOpen=!0)},handleExpiredAtOpenOpenChange:function(t){this.expiredAtOpen=t}}},d=c,l=r("0c7c"),m=Object(l["a"])(d,a,n,!1,null,"9a0d1722",null);e["default"]=m.exports},c8d2:function(t,e,r){var a=r("d039"),n=r("5899"),i="​᠎";t.exports=function(t){return a((function(){return!!n[t]()||i[t]()!=i||n[t].name!==t}))}}}]);