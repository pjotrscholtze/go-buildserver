

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


if (window.location.pathname === "/") {
    setInterval(()=>{
        fetch("http://localhost:3000/api/jobs",{
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            }}).then((r)=>{return r.json()}).
            then((r)=>{
                ensureTableContent(document.querySelector("table"), r);
            })
    }, 300)        
} else if (window.location.pathname.startsWith("/repo/")) {
    setInterval(()=>{
        fetch("http://localhost:3000/api/repos",{
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            }}).then((r)=>{return r.json()}).
            then((r)=>{
                const repoName = document.querySelector(".properties").getAttribute("data-repo");
                const candidates = r.filter((e)=>{return e.Name === repoName});
                if (candidates.length === 0) {
                    // This is weird...
                    return;
                }
                const repoData = candidates[0];
                if (repoData.LastBuildResult.length > 0) {
                    document.querySelector("#build-status span").innerText = repoData.LastBuildResult[0].Status;
                    document.querySelector("#build-reason span").innerText = repoData.LastBuildResult[0].Reason;
                    document.querySelector("#build-start-time span").innerText = repoData.LastBuildResult[0].StartTime;
                    const resultsChanged = ensureTableContent(document.querySelector("table"), repoData.LastBuildResult[0].Lines);
                    if (resultsChanged && document.getElementById("auto-scroll").checked) {
                        window.scrollTo(0, document.body.scrollHeight);
                    }
               }
            })
    }, 300)        
}