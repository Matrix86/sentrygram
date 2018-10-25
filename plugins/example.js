function compareByCpu(a,b) {
  if (a.Cpu > b.Cpu)
    return -1;
  if (a.Cpu < b.Cpu)
    return 1;
  return 0;
}

function padEnd(s,l,c){
    pad = "";
    for(j = 0; j < l - s.length; j++) {
        pad += c;
    }
    return s + pad;
}

function GetTopCpuProcs(req, res) {
    procs = getProcesses()
    sorted = procs.sort(compareByCpu);
    result = "<pre> PID      CPU%     Mem%     Name \n";
    for(i = 0; i < 10; i++) {
        pid  = sorted[i].Pid.toString();
        cpu  = (Math.round(sorted[i].Cpu * 100) / 100).toString();
        mem  = (Math.round(sorted[i].Mem * 100) / 100).toString();
        name = sorted[i].Name;
        result += " " + padEnd(pid, 9, " ") + padEnd(cpu, 9, " ") + padEnd(mem, 9, " ") + name + "\n";
    }
    result += "</pre>";
    res.Content = result;
}

function GetCpuUsage(req, res) {
    cpus = cpuUsage();
    label = [];
    for(i=0; i < cpus.length; i++) {
        label[i] = "CPU"+i;
    }

    file = newBarGraph("CPU Usage", cpus, label);
    if( file != "") {
        res.Content = file;
        res.Type = ImageType;
    } else {
        res.Content = "Problem with image creation";
    }
}
