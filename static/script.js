

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
            if (typeof content[idx][heading[headingIdx]] === "object") {
                while (cells[headingIdx].hasChildNodes()) {
                    cells[headingIdx].removeChild(cells[headingIdx].firstChild);
                }
                cells[headingIdx].appendChild(content[idx][heading[headingIdx]]);
                continue;
            }
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
    createWsConnection("ws://"+window.location.host+"/ws/jobs", function(evt) {
        const jobs = JSON.parse(evt.data);
        console.log(jobs);
        ensureTableContent(document.querySelector("table"), jobs);
    });
} else if (window.location.pathname.startsWith("/repo/") && window.location.pathname.endsWith("/live")) {
    const repoName = document.querySelector(".properties").getAttribute("data-repo");
    createWsConnection("ws://"+window.location.host+"/ws/repo-build-live/"+repoName, function(evt) {
        const repoData = JSON.parse(evt.data);

        document.querySelector("#build-status span").innerText = repoData.Status;
        document.querySelector("#build-reason span").innerText = repoData.Reason;
        document.querySelector("#build-start-time span").innerText = repoData.StartTime;
        const resultsChanged = ensureTableContent(document.querySelector("table"), repoData.Lines);
        if (resultsChanged && document.getElementById("auto-scroll").checked) {
            window.scrollTo(0, document.body.scrollHeight);
        }
    });
} else if (window.location.pathname.startsWith("/repo/") && !window.location.pathname.substr(6).includes("/")) {
    const repoName = document.querySelector(".properties").getAttribute("data-repo");

    document.querySelector("#start-build").setAttribute("href", "#");
    document.querySelector("#start-build").onclick = (ev)=>{
        fetch("/api/pipeline/" + repoName + "?reason=web", {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
        });
    };

    createWsConnection("ws://"+window.location.host+"/ws/repo-builds/"+repoName, function(evt) {
        const repoData = JSON.parse(evt.data);
        console.log(repoData);
        const queryString = window.location.search;
        const urlParams = new URLSearchParams(queryString);
        // urlParams.has('product') ? urlParams.get('product') : 'default';
        const pageNumber = 1 * (urlParams.has('pageNumber') ? urlParams.get('pageNumber') : 1);
        const numberOfResultsPerPage = 1 * (urlParams.has('numberOfResultsPerPage') ? urlParams.get('numberOfResultsPerPage') : 10);
        console.log((pageNumber - 1) * numberOfResultsPerPage, pageNumber * numberOfResultsPerPage);

        let start = Math.max((pageNumber - 1) * numberOfResultsPerPage, 0);
        let end = Math.min(pageNumber * numberOfResultsPerPage, repoData.Jobs.length);

        let jobs = repoData.Jobs.reverse().slice(start, end);
        for (let i in jobs) {
            let anchor = document.createElement("a");
            anchor.innerText = "build-" + jobs[i].ID;
            anchor.setAttribute("href", "/build/build-" + jobs[i].ID)
            jobs[i].ID = anchor;
        }


        const resultsChanged = ensureTableContent(document.querySelector("table"), jobs);

        if (resultsChanged && document.getElementById("auto-scroll").checked) {
            window.scrollTo(0, document.body.scrollHeight);
        }
    });
    console.log("ahhahaha");
} else if (window.location.pathname.startsWith("/build/build-")) {
    const buildNumber = 1 * document.querySelector("#build-id").getAttribute("data-build-id");
    const repoName = document.querySelector(".properties").getAttribute("data-repo");
    createWsConnection("ws://"+window.location.host+"/ws/build/"+buildNumber, function(evt) {
        const repoData = JSON.parse(evt.data);
        document.querySelector("#build-status span").innerText = repoData.Status;
        document.querySelector("#build-reason span").innerText = repoData.Reason;
        const resultsChanged = ensureTableContent(document.querySelector("table"), repoData.Lines);
        if (resultsChanged && document.getElementById("auto-scroll").checked) {
            window.scrollTo(0, document.body.scrollHeight);
        }


    });
    console.log("ahhahaha");
}
