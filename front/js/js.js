$(document).ready(function () {
    var reSplit = function () {
        Split(['#videoGrid', '#videoContent'], {
            direction: 'horizontal',
            minSize: 0,
            sizes: [100, 0],
            gutterSize: 8,
            cursor: 'row-resize',
            "onDragStart": function () {
                console.log($('#videoContent').width())
            }
        });
    };
    reSplit();
    var player = plyr.setup()[0].plyr;

    vue = new Vue({
        el: 'body',
        data: {
            videos: []
        },
        methods: {
            getData: function (e) {
                $.get('https://oizgt5pjf8.execute-api.us-east-1.amazonaws.com/prod/aws-vid-transcoder_webService').done(function (data) {
                    vue.$set('videos', data);
                });
            },
            changeVideo: function (e) {
                var d_key = $(e.target).data('d_key');
                var stamp = $(e.target).data('stamp');
                player.source({
                    type: 'video',
                    sources: [{
                        src: "https://s3.amazonaws.com/idrsainput/output/" + stamp + "%23" + d_key + "%23.webm",
                        type: 'video/webm'
                    }],
                    poster: 'https://s3.amazonaws.com/idrsainput/output/' + stamp + '%23' + d_key + '%23_thumb00001.jpg'
                });
                player.play();
            }
        }

    });

    vue.getData();
    window.setInterval(function () {
        vue.getData();
        console.log('Refreshed data');
    }, 30000);


});