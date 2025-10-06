import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Тип нагрузки
export const options = {
    vus: 100,
    duration: '2s',
};

const authorIds = [
    '1430e926-b935-4dd5-b0dc-07b0457149c6',
    'c57ebf06-004b-414e-9f06-76bb3000efc9',
];

export default function () {
    const url = 'http://localhost:8080/v1/library/book';

    const randomAuthorId = randomItem(authorIds);

    const payload = JSON.stringify({
        name: `book-${__VU}-${__ITER}`,
        author_id: [randomAuthorId],
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    });

    sleep(0.1);
}