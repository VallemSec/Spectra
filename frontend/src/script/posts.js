// Description: Fetch all posts from the WordPress REST API at https://ismijnbedrijfveilig.nl/wp-json/wp/v2/posts

const div = document.getElementById('blogposts');

// Get the post data from the WordPress REST API
async function fetchPosts() {
    const requestOptions = {
        method: "GET",
        redirect: "follow"
    };

    try {
        const response = await fetch("https://ismijnbedrijfveilig.nl/wp-json/wp/v2/posts", requestOptions);
        const result = await response.json();
        formatPosts(result);
    } catch (error) {
        console.error(error);
    }
}

fetchPosts();

// Format the post data and display it on the page

function formatPosts(posts) {
    posts.slice(0, 3).forEach(post => {
        const postDiv = document.createElement('div');
        postDiv.className = 'col-span-12 lg:col-span-4';
        const date = new Date(post.date);
        const formattedDate = `${date.getDate()}-${date.getMonth() + 1}-${date.getFullYear()}`;
        postDiv.innerHTML = `
            <a class="post group" href="${post.link}" target="_blank">
            <img class="rounded-md group-hover:scale-105 transition-all" src="${post.jetpack_featured_media_url}">
            <p class="text-sm text-gray-700 mt-2 transition-all">${formattedDate}</p>
            <h2 class="text-xl group-hover:text-blue-vallem transition-colors">${post.title.rendered}</h2>
            </a>
        `;
        div.appendChild(postDiv);
    });
}

console.log(results);
