var activities;

function getActivities() {
    fetch('/api/v1/activities', {
        headers: {
        'content-type': 'application/json'
        }})
        .then(function(response) {
            return response.json();
        })
        .then(function(myJson) {
            activities = myJson;
        });
};

getActivities();

google.charts.load('current', {'packages':['timeline']});
google.charts.setOnLoadCallback(drawChart);

function drawChart() {

    var data = new google.visualization.DataTable();
    data.addColumn('string', 'App Name');
    data.addColumn('string', 'Title');
    data.addColumn('date', 'Start Date');
    data.addColumn('date', 'End Date');

    var rows = activities['activities'].map(x => [x.app_name, x.name, new Date(x.start_time), new Date(x.end_time)])

    data.addRows(rows);

    var options = {
        height: 800,
        colors: ['red', 'blue', 'green', 'orange', 'purple'],
    };

    var chart = new google.visualization.Timeline(document.getElementById('chart_div'));

    chart.draw(data, options);
}
