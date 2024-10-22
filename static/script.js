

function ensureTableContent(table, content) {
    let changeOccured = false;
    let heading = [];
    document.querySelectorAll("thead td").forEach((e)=>{heading.push(e.innerText)})

    // Too many rows in table.
    while (table.querySelectorAll("tbody tr").length > content.length) {
        table.querySelector("tbody").removeChild(table.querySelectorAll("tbody tr")[0]);
        changeOccured = true;
    }

    // Too little rows in table.
    while (table.querySelectorAll("tbody tr").length < content.length) {
        let tr = document.createElement("tr");
        for (let i in heading) {
            let td = document.createElement("td");
            tr.appendChild(td);
        }

        table.querySelector("tbody").appendChild(tr);
        changeOccured = true;
    }

    // Ensure contents of rows.
    let rows = table.querySelectorAll("tbody tr");
    for (let idx = 0; idx < rows.length; idx++) {
        const cells = rows[idx].querySelectorAll("td");
        for (let headingIdx = 0; headingIdx < heading.length; headingIdx++) {
            if (cells[headingIdx].innerText == content[idx][heading[headingIdx]]) {
                continue;
            }
            cells[headingIdx].innerText = content[idx][heading[headingIdx]];
            changeOccured = true;
        }
    }
    return changeOccured;
}

function createWsConnection(endpoint, onMessage) {
    function setupWebsocket() {
        var conn = new WebSocket(endpoint);
        conn.onclose = function(evt) {
            console.log('Connection closed', evt);
            setTimeout(setupWebsocket, 2000);
        }
        conn.onmessage = onMessage;
        conn.onerror = ()=>{
            setTimeout(setupWebsocket, 2000);
        };
        setInterval(()=>{
            try{
                conn.send("ping");
            } catch(Exception){}
        }, 1000);
    }
    setupWebsocket();
}

if (window.location.pathname === "/") {
    createWsConnection("ws://localhost:3000/ws/jobs", function(evt) {
        const jobs = JSON.parse(evt.data);
        console.log(jobs);
        ensureTableContent(document.querySelector("table"), jobs);
    });
} else if (window.location.pathname.startsWith("/repo/")) {
    const repoName = document.querySelector(".properties").getAttribute("data-repo");
    createWsConnection("ws://localhost:3000/ws/repo/"+repoName, function(evt) {
        const repoData = JSON.parse(evt.data);

        document.querySelector("#build-status span").innerText = repoData.Status;
        document.querySelector("#build-reason span").innerText = repoData.Reason;
        document.querySelector("#build-start-time span").innerText = repoData.StartTime;
        const resultsChanged = ensureTableContent(document.querySelector("table"), repoData.Lines);
        if (resultsChanged && document.getElementById("auto-scroll").checked) {
            window.scrollTo(0, document.body.scrollHeight);
        }
    });
}