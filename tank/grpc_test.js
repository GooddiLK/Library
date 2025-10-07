import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    vus: 100,
    duration: '2s',
};

const client = new grpc.Client();
client.load(['../api/library'], 'library.proto');

const authorIds = [
    '0404622f-aa36-481c-b64d-c93f87357ff5'
];

export default function () {
    if (!client.connected) {
        client.connect('localhost:9090', {
            plaintext: true,
        });
    }

    const randomAuthorId = randomItem(authorIds);

    const payload = {
        name: `book-${__VU}-${__ITER}`,
        author_id: [randomAuthorId],
    };

    const res = client.invoke(
        'library.Library/AddBook',
        payload
    );

    check(res, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    sleep(0.1);
}

export function teardown() {
    client.close();
}
