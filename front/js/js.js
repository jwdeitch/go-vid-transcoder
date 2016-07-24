// credit http://stackoverflow.com/a/18650828/4603498
function formatBytes(bytes, decimals) {
    if (bytes == 0) return '0 Byte';
    var k = 1000; // or 1024 for binary
    var dm = decimals + 1 || 3;
    var sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    var i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

$(document).ready(function () {
    var reSplit = function () {
        Split(['#videoGrid', '#videoContent'], {
            direction: 'horizontal',
            minSize: 0,
            sizes: [50, 50],
            gutterSize: 8,
            cursor: 'row-resize',
            "onDragStart": function () {
                console.log($('#videoContent').width())
            }
        });
    };
    reSplit();
    var player;

    vue = new Vue({
        el: 'body',
        data: {
            videos: [],
            video: "NA",
            poster: "NA",
            downloadSize: "NA",
            name: "NA",
            stamp: "NA",
            d_key: "NA",
            initialized: false
        },
        watch: {
            'initialized': function () {
                player = plyr.setup()[0].plyr;
            },
            'd_key': function () {
                player.source({
                    type: 'video',
                    sources: [{
                        src: "https://s3.amazonaws.com/idrsainput/output/" + vue.$get('stamp') + "%23" + vue.$get('d_key') + "%23.webm",
                        type: 'video/webm'
                    }],
                    poster: 'https://s3.amazonaws.com/idrsainput/output/' + vue.$get('stamp') + '%23' + vue.$get('d_key') + '%23_thumb00001.jpg'
                });
                player.play();
            }
        },
        methods: {
            thumbnailScroll: function(e) {
                var target = $(e.target);
                var d_key = target.data('d_key');
                var stamp = target.data('stamp');
                var tcount = target.data('tcount');
                x = e.pageX - target.offset().left;
                y = e.pageY - target.offset().top;
                thumbToShow = ((x/target.width())*tcount);
            },
            getData: function (e) {
                $.get('https://oizgt5pjf8.execute-api.us-east-1.amazonaws.com/prod/aws-vid-transcoder_webService').done(function (data) {
                    vue.$set('videos', data);
                });
            },
            handelThumbClick: function (e) {
                if (!vue.$get('initialized')) {
                    vue.$set('initialized', true);
                }
                vue.changeVideo(e);
            },
            changeVideo: function (e) {
                vue.$set('d_key', $(e.target).data('d_key'));
                vue.$set('stamp', $(e.target).data('stamp'));
                vue.$set('name', $(e.target).data('name'));
                vue.$set('downloadSize', formatBytes($(e.target).data('size'), 1));
            }
        }

    });

    vue.getData();
    window.setInterval(function () {
        vue.getData();
        console.log('Refreshed data');
    }, 30000);


});