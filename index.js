const apiConfig = {
    url: `/upload/midi`,
    method: 'POST',
    mode: 'cors'
};
function getApiParams(payload) {
    let { method, mode } = apiConfig;
    return {
        method,
        body: payload,
        mode
    }

}

function getFetchArgs(payload) {
    return [apiConfig.url, getApiParams(payload)];
}

async function awaitFetch(url, params) {
    try {
        let res = await window.fetch(url, params);
        return await res.json();
    } catch (e) {
        console.error(e);
        throw e;
    }
}

function displayDownloadButon(url) {
    const btn = document.querySelector('#download-file');
    btn.addEventListener('click', e => window.location = url);
    btn.classList.remove('hide');
}


async function submitUploadClick(e) {
    const formEl = document.querySelector("#file-upload");
    const outputNameEl = formEl.querySelector("#output-name");
    const waveFormEl = formEl.querySelector("#waveform");
    const fileInputEl = formEl.querySelector("#upload-file");
    const files = fileInputEl.files;
    const formData = new FormData();
    formData.append('myMIDIFile', files[0]);
    formData.append(outputNameEl.name, outputNameEl.value);
    formData.append(waveFormEl.name, waveFormEl.value);
    let data = await awaitFetch(...getFetchArgs(formData));
    console.log('API response from /upload/midi...');
    console.dir(data);
    if (data) {
        displayDownloadButon(data.url);
    }
}

document.querySelector('#file-upload-submit').addEventListener('click', submitUploadClick);