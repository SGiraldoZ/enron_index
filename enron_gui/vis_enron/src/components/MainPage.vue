<script setup>
import { ref, watch } from 'vue';

// Define reactive variables
const searchQuery = ref('');
const mails = ref(null);
const curEmail = ref(null);
const totalPages = ref(1);
const currentPage = ref(1);
const pageResults = ref(20);

// Function to fetch emails
const fetchEmails = async () => {
    try {
        const queryParams = {
            query: searchQuery.value || '',
            n_from: (currentPage.value - 1) * pageResults.value,
            max_results: pageResults.value || '',
        }

        const queryString = Object.entries(queryParams)
            .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
            .join('&');

        const response = await fetch(`/api/enron?${queryString}`);
        // const response = await fetch(`/api/enron`);
        const data = await response.json();
        console.log("Fetching emails");
        console.log(data);
        totalPages.value = Math.ceil(data.hits.total.value / pageResults.value);
        updateEmails(data.hits);
    } catch (error) {
        console.error('Error fetching emails:', error);
    }

};


// Function to update emails data
const updateEmails = (newEmails) => {
    mails.value = newEmails;
};

//SELECT AND RENDER AN EMAIL
const handleEmailClick = (index) => {
    // Handle email click logic here
    console.log("CHanging the value", index)
    curEmail.value = index;
};

//Func to save search value and submit request
const submitSearch = (newQuery) => {
    searchQuery.value = newQuery;
    fetchEmails();
};

// Function to go to the previous page
const prevPage = () => {
    currentPage.value -= 1;
};

// Function to go to the next page
const nextPage = () => {
    currentPage.value += 1;
};

// Watch for changes in the currentPage variable and fetch emails accordingly
watch(currentPage, () => {
    fetchEmails();
});

watch(pageResults, () => {
    fetchEmails();
});

// Initial fetch on component mount
fetchEmails();
</script>

<template>
    <div class="flex flex-col py-2">
        <div>
            <!-- <div class="relative inset-x-0 top-0 h-16"> -->
            <div class="relative mt-2 rounded-md ">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                    <span class="text-gray-500 sm:text-sm">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 50 24" fill="currentColor" class="w-6 h-6">
                            <path fill-rule="evenodd"
                                d="M10.5 3.75a6.75 6.75 0 1 0 0 13.5 6.75 6.75 0 0 0 0-13.5ZM2.25 10.5a8.25 8.25 0 1 1 14.59 5.28l4.69 4.69a.75.75 0 1 1-1.06 1.06l-4.69-4.69A8.25 8.25 0 0 1 2.25 10.5Z"
                                clip-rule="evenodd" />
                        </svg>
                    </span>
                </div>
                <input type="text" name="price" id="price"
                    class="block w-full rounded-md border-0 py-1.5 pl-7 pr-20 text-gray-900 ring-1 ring-inset ring-white-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    placeholder="search" @keyup.enter="fetchEmails" v-model="searchQuery" />
            </div>
        </div>

        <!-- Screen body -->
        <div class="flex py-8">
            <div class="w-1/2">
                <!-- Pagination and total results widget -->
                <div class="flex items-center justify-between mt-4">
                    <!-- <div>Total results: {{ totalResults }}</div> -->
                    <div>
                        <button @click="prevPage" :disabled="currentPage === 1">Previous</button>
                        Page {{ currentPage }} of {{ totalPages }}
                        <button @click="nextPage" :disabled="currentPage >= totalPages">Next</button>
                    </div>
                    <div>
                        Results in this page:
                        <select v-model="pageResults">
                            <option value="10">10</option>
                            <option value="20">20</option>
                            <option value="50">50</option>
                            <option value="100">100</option>
                        </select>
                    </div>
                </div>

                <div class="relative overflow-x-auto">
                    <table class="w-full text-sm text-left rtl:text-right text-gray-800">
                        <thead class="text-xs text-gray-700 uppercase bg-gray-200">
                            <tr>
                                <th scope="col" class="px-6 py-3">
                                    Subject
                                </th>
                                <th scope="col" class="px-6 py-3">
                                    From
                                </th>
                                <th scope="col" class="px-6 py-3">
                                    To
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <!-- <tr class="bg-white border-b" v-for="(mail,index) in mails.hits" onclick="curEmail=index"> -->
                            <tr :class="{ 'bg-sky-100/80': index === curEmail, 'bg-white': index !== curEmail }"
                                class="border-b  hover:bg-sky-100" v-for="(mail, index) in mails?.hits" :key="index"
                                @click="handleEmailClick(index)">
                                <th scope="row" class="px-6 py-4 font-medium whitespace-nowrap">
                                    {{ mail._source.Subject }}
                                </th>
                                <td class="px-6 py-4">
                                    {{ mail._source.From }}
                                </td>
                                <td class="px-6 py-4">
                                    {{ mail._source.To }}
                                </td>
                            </tr>
                        </tbody>
                    </table>

                </div>

            </div>

            <div class="w-1/2 px-4">
                <div class="text-1xl font-bold">
                    <span>
                        {{ mails?.hits[curEmail]?._source.Subject }}
                    </span>
                </div>

                <div class="text-justify whitespace-pre-line">
                    {{ mails?.hits[curEmail]?._source.Body }}
                </div>
            </div>
        </div>
    </div>
</template>

