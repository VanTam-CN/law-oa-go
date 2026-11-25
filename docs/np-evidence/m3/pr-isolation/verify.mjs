// Run from the repository root. Only YAML and Git objects are read.
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {readFileSync, readdirSync} from 'node:fs';
import {Lexer, Parser, Evaluator, data} from '@actions/expressions';
import {parse} from 'yaml';

const baseline = '34099b75ddc5d6cc276a51a47a1f157065842e1f';
const m2 = 'codex/np-main-m2', m3 = 'codex/np-main-m3';
const dir = '.github/workflows';
const files = readdirSync(dir).filter(f => /\.ya?ml$/.test(f)).sort();
const namesAt = ref => execFileSync('git', ['ls-tree','--name-only', ref + ':' + dir], {encoding:'utf8'})
  .trim().split('\n').filter(f => /\.ya?ml$/.test(f)).sort();
assert.deepEqual(files, namesAt(baseline));
function read(ref) {
  return Object.fromEntries(files.map(f => [f, parse(ref
    ? execFileSync('git',['show',ref + ':' + dir + '/' + f],{encoding:'utf8'})
    : readFileSync(dir + '/' + f,'utf8'))]));
}
const old = read(baseline), now = read(process.argv[2]);
const phases = ['success','failure','cancelled','failure+cancelled'];
function evaluate(raw, head, phase, event='pull_request') {
  const ctx = {github:{head_ref:head, base_ref:event === 'pull_request' ? 'main' : '',
    ref:event === 'pull_request' ? 'refs/pull/1/merge' : 'refs/heads/main',
    event_name:event, event:{inputs:{}}}, inputs:{}, matrix:{probe:'normal'},
    needs:{preflight:{outputs:{test_type:'all'}}}, env:{SLACK_WEBHOOK_URL:'configured',
      EMAIL_USERNAME:'configured',EMAIL_PASSWORD:'configured',EMAIL_FROM:'configured',NOTIFICATION_EMAIL:'configured'}};
  const funcs = ['success','failure','cancelled','always'].map(name => ({
    name, minArgs:0, maxArgs:0, call:()=>new data.BooleanData(name === 'always' || phase.split('+').includes(name))
  }));
  const expr=raw.trim().replace(/^\$\{\{([\s\S]*)\}\}$/, '$1');
  const ast=new Parser(new Lexer(expr).lex().tokens,Object.keys(ctx),funcs).parse();
  return new Evaluator(ast,JSON.parse(JSON.stringify(ctx),data.reviver),new Map(funcs.map(f=>[f.name,f]))).evaluate().coerceString();
}

// Compare every parsed field, including event filters, needs, commands, action versions and inputs.
// Only scalar branch-dependent guards/settings may differ, and their evaluated values must match.
function effective(tree, head, phase, event) {
  return JSON.parse(JSON.stringify(tree,(key,value)=>{
    if (typeof value !== 'string' || !value.includes('github.head_ref')) return value;
    assert(['if','cache','PILOT_PR'].includes(key), 'inspect new branch-dependent field: '+key);
    return evaluate(value,head,phase,event);
  }));
}
function allowed(condition, head, phase) {
  // Over-approximate step/job reachability: a successful sibling or an earlier step
  // can still have run even if the workflow ultimately failed or was cancelled.
  return !condition || evaluate(condition,head,phase).toLowerCase()==='true';
}
function surfaces(tree,head,phase) {
  const contacts=[], jobs=[];
  for (const [file,wf] of Object.entries(tree)) {
    assert(wf.on.pull_request.branches.includes('main'));
    const cache=new Map();
    function reachable(id) {
      if (!cache.has(id)) {
        const j=wf.jobs[id], parents=[].concat(j.needs??[]);
        const override=/\b(always|failure|cancelled|success)\(/i.test(j.if??'');
        cache.set(id, (override || parents.every(reachable)) && phases.some(p=>allowed(j.if,head,p)));
      }
      return cache.get(id);
    }
    for (const [id,j] of Object.entries(wf.jobs)) {
      if (!reachable(id)) continue;
      jobs.push(file+'/'+id);
      if (id.startsWith('deploy-')) contacts.push(file+'/'+id+' deployment');
      for (const s of j.steps??[]) {
        if (!s.uses || !allowed(s.if,head,phase)) continue;
        const name=s.uses.split('@')[0];
        if (name==='actions/checkout') continue;
        if (['actions/setup-go','actions/setup-node'].includes(name)) {
          const setting=s.with?.cache??(name==='actions/setup-go');
          const value=String(setting).includes('${{') ? evaluate(setting,head,phase) : String(setting);
          if (!value || value==='false') continue;
          contacts.push(file+'/'+id+' tool-cache');
        } else contacts.push(file+'/'+id+' '+name);
      }
    }
  }
  return {jobs,contacts};
}
const result={baseline,profiles:[],negative_controls:[]};
for (const head of [m2,m3,'feature/example',m3+'-followup','codex/np-main-m30']) {
  for (const phase of phases) {
    const pilot=head===m2 || head===m3;
    assert.deepEqual(effective(now,head,phase,'pull_request'),effective(old,pilot?m2:head,phase,'pull_request'),
      head+' '+phase+' effective workflow differs from reference');
    const {jobs,contacts}=surfaces(now,head,phase);
    if (pilot) {
      assert.deepEqual(contacts,[]);
      for (const job of ['tests.yml/np-m2-early','pgo.yml/pilot-pgo','tests.yml/e2e-tests','ci.yml/quality']) assert(jobs.includes(job));
    } else {
      assert(!jobs.includes('tests.yml/np-m2-early'));
      assert.deepEqual(contacts,surfaces(old,head,phase).contacts);
      assert(contacts.some(c=>c.endsWith('action-slack')));
    }
    result.profiles.push({head,phase,eligible_jobs:jobs.length,possible_external_steps:contacts.length});
  }
}
for (const event of ['push','workflow_dispatch','schedule','release']) {
  for (const p of phases) assert.deepEqual(effective(now,'',p,event),effective(old,'',p,event));
}
const bad=structuredClone(now);
bad['tests.yml'].jobs['e2e-tests'].steps.find(s=>s.with?.name==='playwright-screenshots').if='failure() || cancelled()';
bad['ci.yml'].jobs.notify.if='always()';
bad['ci.yml'].jobs.quality.steps.find(s=>s.uses?.startsWith('actions/setup-go')).with.cache=true;
for (const head of [m2,m3]) for (const phase of ['failure','cancelled']) {
  const {contacts}=surfaces(bad,head,phase);
  for (const marker of ['e2e-tests actions/upload-artifact','quality tool-cache','notify 8398a7/action-slack','notify dawidd6/action-send-mail'])
    assert(contacts.some(c=>c.endsWith(marker)), marker);
  result.negative_controls.push({head,phase,detected:contacts});
}
console.log(JSON.stringify(result,null,2));
