from flask import Flask, request, jsonify
import time
import random

app = Flask(__name__)

@app.route('/pay', methods=['POST'])
def pay():
    data = request.json
    mode = data.get('mode', 'happy')
    
    if mode == 'timeout':
        time.sleep(10)  # Simulate timeout
        return jsonify({"status": "timeout"}), 504
    elif mode == 'error':
        return jsonify({"status": "error"}), 500
    elif mode == 'declined':
        return jsonify({"status": "declined"}), 400
    elif mode == 'latency':
        time.sleep(random.uniform(1, 3))
        return jsonify({"status": "approved"}), 200
    elif mode == 'random':
        outcomes = ['approved', 'declined', 'error', 'timeout']
        status = random.choice(outcomes)
        if status == 'timeout':
            time.sleep(10)
            return jsonify({"status": status}), 504
        elif status == 'error':
            return jsonify({"status": status}), 500
        elif status == 'declined':
            return jsonify({"status": status}), 400
        else:
            return jsonify({"status": status}), 200
    else:  # happy
        return jsonify({"status": "approved"}), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)