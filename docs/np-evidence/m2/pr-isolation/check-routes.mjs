import assert from 'node:assert/strict';
import {readFileSync, readdirSync} from 'node:fs';
import {Parser, Lexer, Evaluator, data} from '@actions/expressions';
import {parse} from 'yaml';

const dir = '.github/workflows';
const files = readdirSync(dir).filter(f => f.endsWith('.yml')).sort();
assert.equal(files.length, 6);
const workflows = Object.fromEntries(files.map(f => [f, parse(readFileSync(dir+'/'+f, 'utf8'))]));
const pilot = 'codex/np-main-m2';
const statuses = ['success', 'failure', 'cancelled'];
const isStatus = /\b(?:always|success|failure|cancelled)\(/i;
function values(raw, context) {
  // Include mid-job failure/cancellation even when every upstream job succeeded.
  return statuses.map(status => {
    const functions = ['success','failure','cancelled','always'].map(name => ({
      name, minArgs:0, maxArgs:0, call:() => new data.BooleanData(name === 'always' || name === status)
    }));
    const ast = new Parser(new Lexer(String(raw)).lex().tokens, Object.keys(context), functions).parse();
    const result = new Evaluator(ast, JSON.parse(JSON.stringify(context), data.reviver),
      new Map(functions.map(f => [f.name, f]))).evaluate();
    return result.coerceString();
  });
}
function possible(raw, context) {
  return !raw || values(raw,context).some(v=>v.toLowerCase()==='true');
}
function audit(event, head, docs=workflows) {
  const contacts=[], jobs=[];
  for (const [file, wf] of Object.entries(docs)) {
    assert(wf.on[event].branches.includes('main'));
    const reachable = new Map();
    const ctx = {github:{event_name:event,head_ref:head,ref:event==='push'?'refs/heads/main':'refs/pull/1/merge',event:{inputs:{}}},
      inputs:{},env:{SLACK_WEBHOOK_URL:'configured',EMAIL_USERNAME:'configured',EMAIL_PASSWORD:'configured',EMAIL_FROM:'configured',NOTIFICATION_EMAIL:'configured'},
      needs:{preflight:{outputs:{test_type:'all'}}},matrix:{}};
    function canRun(id) {
      if (!reachable.has(id)) {
        const job = wf.jobs[id];
        const blocked = [].concat(job.needs??[]).some(parent => !canRun(parent));
        reachable.set(id, (!blocked || isStatus.test(job.if??'')) && possible(job.if,ctx));
      }
      return reachable.get(id);
    }
    for (const [id,job] of Object.entries(wf.jobs)) {
      if (!canRun(id)) continue;
      jobs.push(file+'/'+id);
      for (const step of job.steps || []) {
        const uses = step.uses?.split('@')[0];
        if (uses === 'actions/setup-go' || uses === 'actions/setup-node') {
          const setting=step.with?.cache ?? (uses === 'actions/setup-go');
          const expression=String(setting).match(/^\$\{\{(.*)\}\}$/);
          const resolved=expression ? values(expression[1],ctx) : [String(setting)];
          if (resolved.some(v=>v && v.toLowerCase()!=='false') && possible(step.if,ctx)) contacts.push([file,id,'tool-cache',step.name]);
        }
        if (uses && !['actions/checkout','actions/setup-go','actions/setup-node'].includes(uses)) {
          if (possible(step.if,ctx)) contacts.push([file,id,uses,step.name]);
        }
      }
      if (id.startsWith('deploy-')) contacts.push([file,id,'deployment',job.name]);
    }
  }
  return {jobs,contacts};
}
const actual=audit('pull_request',pilot);
assert.deepEqual(actual.contacts,[]);
for (const id of ['tests.yml/np-m2-early','tests.yml/go-tests','tests.yml/frontend-tests','tests.yml/e2e-tests','pgo.yml/pilot-pgo','security.yml/security-smoke','ci.yml/quality']) assert(actual.jobs.includes(id), id);
const ordinary=audit('pull_request','feature/example'), push=audit('push','');
assert(ordinary.contacts.some(row=>row[2]==='codecov/codecov-action'));
assert(ordinary.contacts.some(row=>row[2]==='actions/github-script'));
assert(push.contacts.some(row=>row[2]==='docker/build-push-action'));
const mutated=structuredClone(workflows);
mutated['tests.yml'].jobs['e2e-tests'].steps.find(s=>s.with?.name==='playwright-screenshots').if='failure()';
mutated['ci.yml'].jobs.notify.if='always()';
mutated['ci.yml'].jobs.quality.steps.find(s=>s.uses?.startsWith('actions/setup-go')).with.cache=true;
const restored=audit('pull_request',pilot,mutated).contacts;
assert(restored.some(row=>row[1]==='e2e-tests' && row[2]==='actions/upload-artifact'));
assert(restored.some(row=>row[2]==='tool-cache'));
assert(restored.some(row=>row[2]==='8398a7/action-slack'));
assert(restored.some(row=>row[2]==='dawidd6/action-send-mail'));
console.log(JSON.stringify({files,pilot_jobs:actual.jobs,pilot_external_steps:actual.contacts,
  ordinary_pr_external_step_count:ordinary.contacts.length,main_push_external_step_count:push.contacts.length,
  mutation_detected:restored.map(row=>row.slice(0,3))},null,2));
