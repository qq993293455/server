(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-07cb459b"],{9406:function(t,e,n){"use strict";n.r(e);var o=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("dash-board",{attrs:{"is-edit":!1}})},r=[],c=(n("caad"),n("2532"),function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"components-container"},[n("el-form",{ref:"form",staticStyle:{width:"50%"},attrs:{model:t.form,"label-width":"120px"}},[n("el-form-item",{attrs:{label:"选择share目录"}},[n("el-input",{staticClass:"input-with-select",attrs:{placeholder:"请输入目录"},model:{value:t.form.workdir,callback:function(e){t.$set(t.form,"workdir",e)},expression:"form.workdir"}},[n("el-button",{attrs:{slot:"append",icon:"el-icon-circle-plus",type:"primary"},on:{click:t.setWorkdir},slot:"append"},[t._v(" save ")])],1)],1),n("el-form-item",{attrs:{label:"是否接收push"}},[n("el-switch",{attrs:{"active-color":"#13ce66","inactive-color":"#ff4949"},on:{change:t.pushChange},model:{value:t.form.push,callback:function(e){t.$set(t.form,"push",e)},expression:"form.push"}})],1),n("el-form-item",{attrs:{label:"导入规则中心"}},[n("el-button",{attrs:{icon:"el-icon-circle",type:"primary"},on:{click:t.genRule}},[t._v(" GO ")])],1),n("el-form-item",{attrs:{label:"生成proto"}},[n("el-button",{attrs:{icon:"el-icon-circle",type:"primary"},on:{click:t.genProto}},[t._v(" GO ")])],1)],1)],1)}),a=[],i=n("b775");function s(t){return Object(i["a"])({url:"/workdir/save",method:"post",dir:t,transformRequest:[function(){var e="";return e+=encodeURIComponent("dir")+"="+encodeURIComponent(t)+"&",e=e.substring(0,e.lastIndexOf("&")),e}],headers:{"Content-Type":"application/x-www-form-urlencoded"}})}function u(){return Object(i["a"])({url:"/workdir/get",method:"get"})}function l(){return Object(i["a"])({url:"/push/get",method:"get"})}function f(t){return Object(i["a"])({url:"/push/set",method:"post",push:t,transformRequest:[function(){var e="";return e+=encodeURIComponent("push")+"="+encodeURIComponent(t)+"&",e=e.substring(0,e.lastIndexOf("&")),e}],headers:{"Content-Type":"application/x-www-form-urlencoded"}})}function d(){return Object(i["a"])({url:"/gen/rule",method:"get"})}function h(){return Object(i["a"])({url:"/gen/proto",method:"get"})}var p={name:"DashBoard",data:function(){return{form:{push:!0,workdir:""}}},created:function(){this.getWorkdir(),this.getPush()},methods:{setWorkdir:function(){s(this.form.workdir).then((function(t){})).catch((function(t){console.log(t)}))},getPush:function(){var t=this;l().then((function(e){t.form.push=e.data})).catch((function(t){console.log(t)}))},pushChange:function(){f(this.form.push).then((function(t){})).catch((function(t){console.log(t)}))},getWorkdir:function(){var t=this;u(this.form.workdir).then((function(e){t.form.workdir=e.data})).catch((function(t){console.log(t)}))},genRule:function(){d().then((function(t){})).catch((function(t){console.log(t)}))},genProto:function(){h().then((function(t){})).catch((function(t){console.log(t)}))}}},m=p,b=n("2877"),g=Object(b["a"])(m,c,a,!1,null,"56ff6d86",null),w=g.exports,k={name:"Dashboard",components:{DashBoard:w},data:function(){return{currentRole:"adminDashboard"}},created:function(){this.roles.includes("admin")||(this.currentRole="editorDashboard")}},v=k,O=Object(b["a"])(v,o,r,!1,null,null,null);e["default"]=O.exports},b775:function(t,e,n){"use strict";var o=n("bc3a"),r=n.n(o),c=r.a.create({baseURL:"http://localhost:8080",timeout:5e3});e["a"]=c}}]);