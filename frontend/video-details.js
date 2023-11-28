// video-details.js

document.addEventListener('DOMContentLoaded', function () {
    const videoDetailsContainer = document.getElementById('videoDetails');

    // Extract parameters from the URL
    const params = new URLSearchParams(window.location.search);
    // const videoId = params.get('id');
    const videoUrl = decodeURIComponent(params.get('url'));


    const video = document.getElementById('video');

        if (Hls.isSupported()) {
            const hls = new Hls();
            hls.loadSource(videoUrl);
            hls.attachMedia(video);
        } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = videoUrl;
        }
        // Function to toggle fullscreen
    const toggleFullscreen = () => {
        if (video.requestFullscreen) {
            video.requestFullscreen();
        } else if (video.mozRequestFullScreen) {
            video.mozRequestFullScreen();
        } else if (video.webkitRequestFullscreen) {
            video.webkitRequestFullscreen();
        } else if (video.msRequestFullscreen) {
            video.msRequestFullscreen();
        }
    };
    
    document.addEventListener('DOMContentLoaded', function () {
    video.addEventListener('click', toggleFullscreen);
    
    });
});