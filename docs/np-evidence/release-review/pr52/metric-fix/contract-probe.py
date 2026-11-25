"""Independent process per import topology; no DB or HTTP traffic."""
import json, pathlib, subprocess, sys, tempfile
root=pathlib.Path(sys.argv[1]).resolve()
phase=sys.argv[2]
template=r'''
package main
import (
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "github.com/prometheus/client_golang/prometheus"
 IMPORTS
)
func registered() (*prometheus.HistogramVec, bool) {
 probe := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name:"db_connection_wait_seconds",Help:"Time spent waiting for database connections",Buckets:prometheus.DefBuckets}, []string{"database"})
 err := prometheus.Register(probe)
 if err == nil {
  if !prometheus.Unregister(probe) {panic("temporary probe not removed")}
  return nil,false
 }
 var already prometheus.AlreadyRegisteredError
 if !errors.As(err,&already) {panic(err)}
 hist,ok:=already.ExistingCollector.(*prometheus.HistogramVec)
 if !ok {panic("not a HistogramVec")}
 return hist,true
}
func main() {
 prior,before:=registered()
 if before!=EXPECTED {panic("incorrect registration timing")}
 if before {prior.WithLabelValues("synthetic").Observe(.125)}
 INIT
 hist,after:=registered()
 if !after {panic("missing existing histogram")}
 expectedCount:=uint64(1);expectedSum:=.25
 if before {if prior!=hist {panic("collector replaced during init")};expectedCount++;expectedSum+=.125}
 hist.WithLabelValues("synthetic").Observe(.25)
 families,err:=prometheus.DefaultGatherer.Gather()
 if err!=nil {panic(err)}
 count:=0
 for _,f:=range families {
  if f.GetName()!="db_connection_wait_seconds" {continue}
  count++
  if f.GetHelp()!="Time spent waiting for database connections" || f.GetType().String()!="HISTOGRAM" || len(f.Metric)!=1 {panic("metric family changed")}
  m:=f.Metric[0]; h:=m.GetHistogram()
  if len(m.Label)!=1 || m.Label[0].GetName()!="database" || m.Label[0].GetValue()!="synthetic" || h.GetSampleCount()!=expectedCount || h.GetSampleSum()!=expectedSum || len(h.Bucket)!=len(prometheus.DefBuckets) {panic("metric labels or observations changed")}
  for i,b:=range h.Bucket {if b.GetUpperBound()!=prometheus.DefBuckets[i] {panic("metric bucket changed")}}
 }
 if count!=1 {panic(fmt.Sprint("families ",count))}
 json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"before_init":before,"after_init":after,"families":count,"bucket_count":len(prometheus.DefBuckets),"sample_count":expectedCount, "sample_sum":expectedSum})
}
'''
observed={}
for mode in ("database","monitoring","combined"):
 imports=[]
 if mode!="monitoring": imports+=['_ "law-oa-go/internal/database"']
 if mode!="database": imports+=['"law-oa-go/internal/monitoring"']
 script=template.replace("IMPORTS","\n".join(imports)).replace("EXPECTED",str(mode!="monitoring").lower()).replace("INIT","" if mode=="database" else "monitoring.InitPerformanceMetrics(); monitoring.InitPerformanceMetrics()")
 with tempfile.TemporaryDirectory(prefix="contractprobe",dir=root) as place:
  pathlib.Path(place,"main.go").write_text(script)
  result=subprocess.run(["go","run","./"+pathlib.Path(place).name],cwd=root,capture_output=True,text=True)
 if phase=="before" and mode=="combined":
  assert result.returncode!=0 and "duplicate metrics collector registration attempted" in result.stderr,mode
  observed[mode]={"expected_duplicate_panic":True}
 else:
  if result.returncode:sys.exit(mode+": "+result.stderr[-900:])
  observed[mode]=json.loads(result.stdout.splitlines()[-1])
print(json.dumps(observed))
