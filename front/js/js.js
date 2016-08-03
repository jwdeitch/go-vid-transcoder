// credit http://stackoverflow.com/a/18650828/4603498
function formatBytes(bytes, decimals) {
    if (bytes == 0) return '0 Byte';
    var k = 1000; // or 1024 for binary
    var dm = decimals + 1 || 3;
    var sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    var i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

config = {
    webserviceLambda: "https://oizgt5pjf8.execute-api.us-east-1.amazonaws.com/prod/aws-vid-transcoder_webService"
};

$(document).ready(function () {
    // Dropzone

    var rangeSlider = document.getElementById('thumbnailSizeRange');

    $('.uploadBtn').click(function () {
        $('.dz')[0].click()
    });

    $('.settingsBtn').popup({
        inline: true,
        hoverable: true,
        position: 'bottom center',
        delay: {
            show: 20,
            hide: 500
        }
    });

    var app = {};
    app.signS3RequestURL = 'https://oizgt5pjf8.execute-api.us-east-1.amazonaws.com/prod/aws-vid-transcoder_policy';
    app.S3_BUCKET = 'https://s3.amazonaws.com/idrsainput/';
    var dz = new Dropzone(".dz", {
        url: app.S3_BUCKET,
        method: "post",
        autoProcessQueue: true,
        clickable: true,
        maxfiles: 5,
        parallelUploads: 3,
        maxFilesize: 42950, // 5gb in mb
        maxThumbnailFilesize: 8,
        thumbnailWidth: 150,
        thumbnailHeight: 150,
        acceptedMimeTypes: "video/*",
        previewTemplate: '<div class="dz-preview dz-file-preview"><div class="dz-details"><div class="dz-filename"><b><span data-dz-name></span></b></div><div class="dz-size" data-dz-size></div><div class="uploadBar"><span class="uploadProgress fileUploading" data-dz-uploadprogress>Uploading...</span></div></div></div>',
        accept: function (file, done) {
            $('.uploadBtn').popup({
                hoverable: true,
                duration: 20,
                position: 'bottom center'
            }).popup('show');
            $.ajax({
                async: false,
                url: app.signS3RequestURL,
                dataType: 'json',
                data: JSON.stringify({file: file.name, type: file.type}),
                type: 'POST',
                success: function jQAjaxSuccess(response) {
                    file.policy = response.params;
                    getFileElement(file).addClass(response.params.key.split('#%#')[1].split('.')[0]);
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

    dz.on('success', function (file, response) {
        getFileUploadElement(file).children(".uploadProgress").html("Transcoding    <div class='ui mini active inverted inline loader'></div>").addClass('fileTranscoding').removeClass('fileUploading');
    });

    function getFileUploadElement(file) {
        return $(file.previewTemplate.childNodes[0].childNodes[2]);
    }

    function getFileElement(file) {
        return $(file.previewTemplate.childNodes[0]);
    }


    var reSplit = function () {
        Split(['#videoGrid', '#videoContent'], {
            direction: 'horizontal',
            sizes: [50, 50],
            gutterSize: 8,
            minSize: 300,
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
            videoQueue: [],
            inSearch: false,
            clickBehavior: 'Queue',
            autoPlayNextVideo: false,
            lastAddedVideoQueue: null,
            thumbnailSize: localStorage.getItem('thumbnailSize') ? localStorage.getItem('thumbnailSize') : 200
        },
        watch: {
            'videoQueue': function () {
                plyr.setup(".r" + this.lastAddedVideoQueue);
                if (vue.clickBehavior == "Overwrite" && vue.autoPlayNextVideo) {
                    $('[data-index="0"]')[0].play();
                } else {
                    $(".r" + this.lastAddedVideoQueue).on('ended', function (e) {
                        if (vue.autoPlayNextVideo) {
                            var currentIndex = $(e.target).data('index');
                            var nextVideo = currentIndex + 1;
                            $('[data-index="' + nextVideo + '"]')[0].play();
                            vue.videoQueue.splice(currentIndex, 1);
                        }
                    });
                }
            }
        },
        methods: {
            secondsToString: function (seconds) {
                return moment.duration(seconds, 'seconds').humanize();
            },
            TimeToFromNow: function (UTCtime) {
                return moment.utc(UTCtime).local().fromNow()
            },
            search: function () {
                that = this;
                window.clearTimeout(window.timeOutId);
                window.timeOutId = window.setTimeout(function () {
                    input = $('.search');

                    if (input.val().length > 0) {
                        vue.$set('inSearch', true);
                        $.ajax({
                            type: "GET",
                            url: 'https://oizgt5pjf8.execute-api.us-east-1.amazonaws.com/prod/aws-vid-transcoder_webService' + '?q=' + input.val(),
                            beforeSend: function () {
                                input.parent().addClass('loading');
                            },
                            success: function (data) {
                                vue.$set('videos', data);
                                input.parent().removeClass('loading');
                            },
                            dataType: 'json'
                        });
                    } else {
                        that.$set('inSearch', false);
                        vue.getData()
                    }
                }, 500);
            },
            padThumbnail: function (thumbToShow) {
                return "00000".substring(0, 5 - thumbToShow.toString().length) + thumbToShow;
            },
            thumbPreviewDefault: function (tCount) {
                return this.padThumbnail(Math.floor(tCount / 2));
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

                target.attr('src', "https://s3.amazonaws.com/idrsainput/output/" + stamp + "%23" + d_key + "%23_thumb" + this.padThumbnail(thumbToShow) + ".jpg")
            },
            clearQueue: function () {
                this.videoQueue = [];
            },
            getData: function () {
                $.get(config.webserviceLambda).done(function (data) {
                    data.map(function (obj) {
                        if (obj.Processing === false) {
                            $('.popup .' + obj.DisplayKey + ' .uploadBar .uploadProgress').html("Done!").addClass('fileDone').removeClass('fileTranscoding');
                        }
                    });
                    if (!vue.$get('inSearch')) {
                        vue.$set('videos', data);
                    }
                });
            },
            getDownloadLink: function (video) {
                return "https://s3.amazonaws.com/idrsainput/" + video.Stamp + "%23" + video.DisplayKey + "%23" + video.Name.split('.')[0] + "%2523%2525%2523" + video.DisplayKey + "." + video.Name.split('.')[1];
            },
            findVideoByKeyInFullList: function (key) {
                return this.videos.find(function (video) {
                    return video.DisplayKey === key;
                });
            },
            findVideoByKeyInQueue: function (key) {
                return this.videoQueue.find(function (video) {
                    return video.DisplayKey === key;
                });
            },
            changeVideo: function (e) {
                var videoToAddToQueue = this.findVideoByKeyInFullList($(e.target).data('d_key'));
                var dkey = videoToAddToQueue.DisplayKey;
                var videoInQueue = this.findVideoByKeyInQueue(dkey);
                videoToAddToQueue.downloadSize = formatBytes(videoToAddToQueue.PreTranscodeSize);
                videoToAddToQueue.downloadLink = this.getDownloadLink(videoToAddToQueue);
                videoToAddToQueue.poster = 'https://s3.amazonaws.com/idrsainput/output/' + videoToAddToQueue.Stamp + '%23' + dkey + '%23_thumb' + this.thumbPreviewDefault(videoToAddToQueue.ThumbCount) + '.jpg';
                videoToAddToQueue.src = "https://s3.amazonaws.com/idrsainput/output/" + videoToAddToQueue.Stamp + "%23" + dkey + "%23.webm";

                if (!videoInQueue) {
                    if (this.clickBehavior == 'Queue') {
                        this.videoQueue.push(videoToAddToQueue);
                    }

                    if (this.clickBehavior == 'Stack') {
                        this.videoQueue.unshift(videoToAddToQueue);
                    }

                    if (this.clickBehavior == 'Overwrite') {
                        this.videoQueue = [];
                        this.videoQueue.push(videoToAddToQueue);
                    }

                    this.lastAddedVideoQueue = dkey;
                }
            }
        }

    });

    noUiSlider.create(rangeSlider, {
        start: [vue.$get('thumbnailSize')],
        range: {
            'min': [120],
            'max': [420]
        }
    }).on('update', function (a) {
        localStorage.setItem('thumbnailSize', parseInt(a[0]));
        vue.$set('thumbnailSize', parseInt(a[0]))
    });

    var refreshInterval;
    $('.autoRefresh').checkbox({
        onChange: function () {
            if (this.checked == true) {
                refreshInterval = window.setInterval(function () {
                    vue.getData();
                    console.log('Refreshed data');
                }, 30000);
                localStorage.setItem('autoRefresh', true);
            } else {
                clearInterval(refreshInterval);
                localStorage.setItem('autoRefresh', false);
            }
        }
    });
    if (localStorage.getItem('autoRefresh') == true || localStorage.getItem('autoRefresh') == null) {
        refreshInterval = window.setInterval(function () {
            vue.getData();
            console.log('Refreshed data');
        }, 30000);
        if (localStorage.getItem('autoRefresh') == null) {
            $('.autoRefresh').checkbox('toggle');
        }
    }
    vue.getData();

    $('.AutoPlay').checkbox({
        onChange: function () {
            if (this.checked == true) {
                localStorage.setItem('autoPlay', true);
                vue.autoPlayNextVideo = true;
                if (vue.videoQueue.length) {
                    $('[data-index="0"] .plyr')[0].plyr.play();
                }
            } else {
                vue.autoPlayNextVideo = false;
                localStorage.setItem('autoPlay', false);
            }
        }
    });

    if (localStorage.getItem('autoPlay') == "true" || localStorage.getItem('autoPlay') == null) {
        $('.AutoPlay').checkbox('check');
        vue.autoPlayNextVideo = true;
    }

    if (localStorage.getItem('autoPlay') == "false") {
        $('.AutoPlay').checkbox('uncheck');
    }

    var clickBehavior = localStorage.getItem('clickBehavior');
    if (clickBehavior) {
        vue.$set('clickBehavior', clickBehavior);
    }

    $('.ui.dropdown').dropdown({
        onChange: function (newBehavior) {
            localStorage.setItem('clickBehavior', newBehavior);
            vue.$set('clickBehavior', newBehavior);
        }
    }).dropdown('set selected', vue.clickBehavior);

});