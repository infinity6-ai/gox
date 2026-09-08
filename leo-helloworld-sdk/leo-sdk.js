(() => {

    class LeoSdk {

        async hellworld(msg) {
            return new Promise(resolve => {
                setTimeout(() => resolve(`reply by leosdk: ${msg}`), 2000)
            })
        }

    }

    window.leosdk = new LeoSdk()

})()