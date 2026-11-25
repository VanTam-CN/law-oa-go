import assert from 'node:assert/strict'
import {execFileSync} from 'node:child_process'
import {readFileSync,readdirSync} from 'node:fs'
import {parse} from 'yaml'
import {Lexer,Parser,Evaluator,data} from '@actions/expressions'
const baseline='0ca6eb2'
const files=readdirSync('.github/workflows').filter(n=>n.endsWith('.yml'))
const statuses=['success','failure','cancelled','failure+cancelled']
const repo='owner/repo', ordinary='feature/example', pilot='codex/np-main-i11', m3='codex/np-main-m3'
function evaluate(raw, head, origin, target, phase) {
 const ctx={github:{repository:repo,head_ref:head,base_ref:target,ref:'refs/pull/2/merge',event_name:'pull_request',event:{pull_request:{head:{repo:{full_name:origin}}}}}}
 const funcs=['always','failure','cancelled','success'].map(name=>({name,minArgs:0,maxArgs:0,call:()=>new data.BooleanData(name==='always'||phase.split('+').includes(name))}))
 const expr=raw.trim().replace(/^\$\{\{([\s\S]+)\}\}$/,'$1')
 const ast=new Parser(new Lexer(expr).lex().tokens,['github'],funcs).parse()
 return new Evaluator(ast,JSON.parse(JSON.stringify(ctx),data.reviver),new Map(funcs.map(f=>[f.name,f]))).evaluate().coerceString()
}
function guards(wf){
 const out={}
 function visit(n,p){for(const [k,v] of Object.entries(n)){
 const path=p+'/'+k
 if(v && typeof v==='object') visit(v,path)
 else if(typeof v==='string' && ['if','cache','PILOT_PR'].includes(k) && v.includes('github.head_ref'))out[path]=v
 }}
 visit(wf.jobs,'');return out
}
let compared=0
for(const file of files){
 const f='.github/workflows/'+file, old=guards(parse(execFileSync('git',['show',baseline+':'+f],{encoding:'utf8'}))), cur=guards(parse(readFileSync(f,'utf8')))
 for(const target of ['main','develop']) for(const phase of statuses) for(const [head,origin] of [[ordinary,repo],[pilot,repo],[pilot,'fork/repo'],[pilot+'-extra',repo],[m3,repo],['codex/np-main-m2',repo]]) {
  for(const [key,expr] of Object.entries(cur)){
   const side=origin===repo && head===pilot
   const actual=evaluate(expr,head,origin,target,phase)
   if(key==='/np-i11/if'){assert.equal(actual,side?'true':'false');continue}
   const before=evaluate(old[key],side && key!=='/np-m2-early/if'?m3:head,origin,target,phase)
   assert.equal(actual,before,file+key+' '+head+' '+phase);compared++
  }
 }
}
console.log('I11_PR_ROUTING_VERIFIED',compared)
