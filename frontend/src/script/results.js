import Chart from 'chart.js/auto'
import Typed from 'typed.js';
import tippy from 'tippy.js';
import 'tippy.js/dist/tippy.css';

// Insert domain name
const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);

document.getElementById('domain').innerHTML = urlParams.get('domain');

// Tippy tooltip for disclaimer
tippy('#disclaimer', {
    content: "Het advies is gegenereerd door AI en kan onjuistheden bevatten. Raadpleeg altijd een expert.",
    placement: "bottom",
    trigger: 'click',
});

(async function() {
  const data = [80,20];

  new Chart(
    document.getElementById('score'),
    {
      type: 'doughnut',
      options: {
        backgroundColor: ['#183CDD', 'transparent'],
        borderColor: 'transparent',
        plugins: {
            legend: {
                display: false
            },
            tooltip: {
                enabled: false
            }
        },
      },
      data: {
        labels: data.map(row => row.year),
        datasets: [
          {
            data: data
          }
        ]
      },
    }
  );
})();

(async function() {
    let results;
    // Get the data depending on the environment
    if(process.env.SPECTRA_ENVIRONMENT == "production"){
        const headers = new Headers();
        headers.append("Content-Type", "application/json");

        const raw = JSON.stringify({
        "target": urlParams.get('domain')
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
            result = results;
        })
        .catch((error) => console.error(error));
    }
    else {
        results = require('../testdata/testdata.json');
    }

    // Display AI Advice
    const typed = new Typed('#ai-advice', {
    strings: [results.advice],
    typeSpeed: 5,
    });

    // Display the results
    let scanresults;

    results.results.forEach(problem => {
        document.getElementById('problems').innerHTML += `<div class="relative col-span-4 lg:col-span-2 p-4 shadow-box border-2 border-gray-300 rounded-md">
        <p class="text-xl font-bold flex">
           `+ problem.name +`
        </p>
        <p>`+ problem.ai_advice +`</p>
    </div>`;
    });

    console.log(results.results);
    console.log(scanresults);
})();
