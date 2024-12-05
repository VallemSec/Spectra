const form = document.getElementById('form');
const domainfield = document.getElementById('domain');
const outputDiv = document.getElementById('output');

form.addEventListener('submit', (e) => {
    e.preventDefault();

    const headers = new Headers();
    headers.append("Content-Type", "application/json");

    const raw = JSON.stringify({
    "target": "https://spectra.sakuracloud.nl/"
    });

    const requestOptions = {
    method: "POST",
    headers: headers,
    body: raw,
    redirect: "follow"
    };

    fetch(process.env.SPECTRA_SCANNER_DOMAIN, requestOptions)
    .then((response) => response.text())
    .then((result) => {
        console.log(result);
        outputDiv.innerHTML = result;
    })
    .catch((error) => console.error(error));
});
