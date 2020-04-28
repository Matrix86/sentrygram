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

function GetTopCpuProcs(req) {
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

    sendMessage(req.From, result)
}

function GetCpuUsage(req) {
    cpus = cpuUsage();
    label = [];
    for(i=0; i < cpus.length; i++) {
        label[i] = "CPU"+i;
    }

    file = newBarGraph("CPU Usage", cpus, label);
    if( file != "") {
        sendImage(req.From, file)
    } else {
        sendMessage(req.From, "Problem with image creation")
    }
}

// Message received on a group
function OnGroupMessage(req) {
    console.log("Message received on group "+req.ChatName+" sent from "+req.From);

    // returning false the command are disabled on groups!
    return false
}

// Message received on a private chat
function OnPrivateChatMessage(req) {
    log("Private message received from "+ req.From)
}

function LeaveChat(req) {
    console.log(req.Content)
    if( req.Content.length != 1) {
        sendMessage(req.From, "one argument requested");
        return
    }
    if( leaveGroup(req.Content[0]) ) {
        sendMessage(req.From, "Bye bye "+req.Content[0]);
    } else {
        sendMessage(req.From, "Can't do that sir :(");
    }
}

function LeaveChatById(req) {
    console.log(req.Content)
    if( req.Content.length != 1) {
        sendMessage(req.From, "one argument requested");
        return
    }
    if( leaveGroupById(req.Content[0]) ) {
        sendMessage(req.From, "Bye bye "+req.Content[0]);
    } else {
        sendMessage(req.From, "Can't do that sir :(");
    }
}

function GetCacheIds(req) {
    cache = getCachedIds();
    txt = "";
    for(var key in cache) {
        txt += key+" => "+cache[key]+"\n";
    }
    sendMessage(req.From, txt);

}