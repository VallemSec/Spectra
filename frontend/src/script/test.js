import Chart from 'chart.js/auto'

(async function() {
  const data = [80,20];

  new Chart(
    document.getElementById('acquisitions'),
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
