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
    reSplit()
});