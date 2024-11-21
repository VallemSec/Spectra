const form = document.getElementById('form');
const outputDiv = document.getElementById('output');

form.addEventListener('submit', (e) => {
    e.preventDefault();

    const headers = new Headers();
    headers.append("Content-Type", "application/json");

    const raw = JSON.stringify({
    "target": "vps.nekoluka.nl"
    });

    const requestOptions = {
    method: "POST",
    headers: headers,
    body: raw,
    redirect: "follow"
    };

    fetch("https://spectra.sakuracloud.nl/", requestOptions)
    .then((response) => response.text())
    .then((result) => {
        console.log(result);
        outputDiv.innerHTML = result;
    })
    .catch((error) => console.error(error));
});
