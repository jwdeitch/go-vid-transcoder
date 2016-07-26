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
    // Dropzone
    var app = {};
    app.signS3RequestURL = 'https://oizgt5pjf8.execute-api.us-east-1.amazonaws.com/prod/aws-vid-transcoder_policy';
    app.S3_BUCKET = 'https://s3.amazonaws.com/idrsainput/';
    var dz = new Dropzone(".dropzone123", {
        url: app.S3_BUCKET,
        method: "post",
        autoProcessQueue: true,
        clickable: true,
        maxfiles: 5,
        parallelUploads: 3,
        maxFilesize: 10, // in mb
        maxThumbnailFilesize: 8, // 3MB
        thumbnailWidth: 150,
        thumbnailHeight: 150,
        acceptedMimeTypes: "video/*",
        accept: function (file, done) {
            console.log("efwfwefwe");
            $.ajax({
                async: false,
                url: app.signS3RequestURL,
                dataType: 'json',
                data: JSON.stringify({file: file.name, type: file.type}),
                type: 'POST',
                success: function jQAjaxSuccess(response) {
                    file.policy = response.params;
                    console.log("policy retrieve");
                    done();
                },
                error: function (response) {
                    file.custom_status = 'rejected';
                    if (response.responseText) {
                        response = JSON.parse(response.responseText);
                    }
                    if (response.message) {
                        done(response.message);
                        return;
                    }
                    console.log("policy retrieve error");
                    done('error preparing the upload');
                }
            });
        }
    });
    dz.on("sending", function (file, xhr, data) {
        $.each(file.policy, function (k, v) {
            data.append(k, v);
        });
    });


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
            downloadLink: "NA",
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
            secondsToString: function (seconds) {
                return moment.duration(seconds, 'seconds').humanize();
            },
            thumbnailScroll: function (e) {
                var target = $(e.target);
                var d_key = target.data('d_key');
                var stamp = target.data('stamp');
                var tcount = target.data('tcount');
                x = e.pageX - target.offset().left;
                y = e.pageY - target.offset().top;
                thumbToShow = Math.ceil((x / target.width()) * tcount);
                if (thumbToShow == 0) {
                    thumbToShow = 1
                }
                paddedThumb = "00000".substring(0, 5 - thumbToShow.toString().length) + thumbToShow;
                target.attr('src', "https://s3.amazonaws.com/idrsainput/output/" + stamp + "%23" + d_key + "%23_thumb" + paddedThumb + ".jpg")
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
                var d_key = $(e.target).data('d_key');
                var stamp = $(e.target).data('stamp');
                var name = $(e.target).data('name');
                vue.$set('d_key', d_key);
                vue.$set('stamp', stamp);
                vue.$set('name', name);
                vue.$set('downloadSize', formatBytes($(e.target).data('size'), 1));

                vue.$set('downloadLink', "https://s3.amazonaws.com/idrsainput/" + stamp + "%23" + d_key + "%23" + name);
            }
        }

    });

    vue.getData();
    window.setInterval(function () {
        vue.getData();
        console.log('Refreshed data');
    }, 30000);


});