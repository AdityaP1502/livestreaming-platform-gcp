const videoCardContainer = document.getElementById('videoList');
let video_http = "http://34.101.36.32:8000/stream";
fetch(video_http)
.then(res => res.json())
.then(data => {
    console.log(data.data.streams);
    data.data.streams.forEach(item => {
        addVideoCard(item);
    })
})
.catch(err => console.log(err));

const addVideoCard = (videoData) => {
    console.log(videoData)

    const videoCardLink = document.createElement('a');
    videoCardLink.className = 'video-card';
    videoCardLink.href = `video-details.html?url=${encodeURIComponent(videoData["stream-url"])}`; // Replace with the actual URL

    const videoCard = document.createElement('div');
    videoCard.className = 'video-card'
    // videoCard.href = `video-details.html?&url=${encodeURIComponent(videoData["stream-url"])}`; // Replace with the actual URL
    videoCard.href = `video-details.html?url=${videoData["stream-url"]}`; // Replace with the actual URL

    videoCard.innerHTML = `
    <img src="${videoData.metadata.thumbnail}" class="thumbnail" alt="Video Thumbnail">
    <div class="info">
        <div class="title">${videoData.metadata.title}</div>
        <div class="username">${videoData.username}</div>
        <div class="date">${videoData.metadata.createdAt}</div>
    </div>
    <a href=${videoCard.href}></a>
`;
videoCardLink.appendChild(videoCard);
videoCardContainer.appendChild(videoCardLink);
};
    
// const makeVideoCard = (data) => {
//     videoCardContainer.innerHTML += `
//     <div class="video" onclick="location.href = 'https://youtube.com/watch?v=${data.id}'">
//         <img src="${data.snippet.thumbnails.high.url}" class="thumbnail" alt="">
//         <div class="content">
//             <img src="${data.channelThumbnail}" class="channel-icon" alt="">
//             <div class="info">
//                 <h4 class="title">${data.snippet.title}</h4>
//                 <p class="channel-name">${data.snippet.channelTitle}</p>
//             </div>
//         </div>
//     </div>
//     `;
// }

// search bar

// const searchInput = document.querySelector('.search-bar');
// const searchBtn = document.querySelector('.search-btn');
// let searchLink = "https://www.youtube.com/results?search_query=";

// searchBtn.addEventListener('click', () => {
//     if(searchInput.value.length){
//         location.href = searchLink + searchInput.value;
//     }
// })


// app.js

// document.addEventListener('DOMContentLoaded', function () {
//     // Function to fetch video data from your API
//     function fetchVideos() {
//         // Replace 'YOUR_API_ENDPOINT' with the actual endpoint of your API
//         fetch('YOUR_API_ENDPOINT')
//             .then(response => response.json()) 
//             .then(data => displayVideos(data))
//             .catch(error => console.error('Error fetching videos:', error));
//     }

//     // Function to display videos on the page
//     function displayVideos(videos) {
//         // Replace '.video-container' with the appropriate selector for your video container
//         const videoContainer = document.querySelector('.video-container');

//         // Clear existing content
//         videoContainer.innerHTML = '';

//         // Iterate through the videos and create HTML elements
//         videos.forEach(video => {
//             const videoElement = document.createElement('div');
//             videoElement.className = 'video';

//             videoElement.innerHTML = `
//                 <img src="${video.thumbnail}" class="thumbnail" alt="">
//                 <div class="content">
//                     <img src="${video.channelIcon}" class="channel-icon" alt="">
//                     <div class="info">
//                         <h4 class="title">${video.title}</h4>
//                         <p class="channel-name">${video.channelName}</p>
//                     </div>
//                 </div>
//             `;

//             // Append the video element to the container
//             videoContainer.appendChild(videoElement);
//         });
//     }

//     // Fetch videos when the page loads
//     fetchVideos();
// });
