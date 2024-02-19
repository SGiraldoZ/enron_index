<script setup>
import { ref } from 'vue';
const mails = ref(null);
fetch('/api/enron')
    .then(response => response.json())
    .then(data => mails.value = data.hits);
let curEmail = ref(null);

const handleEmailClick = (index) => {
    // Handle email click logic here
    console.log("CHanging the value", index)
    curEmail.value = index;
};
</script>

<template>
    <div class="flex">
        <div class="w-1/2">

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
                        <tr :class="{ 'bg-sky-100/80': index === curEmail,'bg-white': index !== curEmail }"
                            class="border-b  hover:bg-sky-100"
                            v-for="(mail, index) in mails?.hits" :key="index" @click="handleEmailClick(index)">
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

        <div class="w-1/2 px-8">
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
</template>


<style>
/* Define the styling for the active row */
.active {
  background-color: #c27ad4; /* Change to the desired active background color */
}
</style>