const fs = require("fs")

function nameToType(name) {
    if (name.startsWith("animal-")) {
        return {
            type: "animal",
            repeat: 1,
            variant: +name.substring("animal-".length)
        }
    } else {
        let fruits = name.split("_").map(seg => {
            return {
                variant: ["s", "p", "g", "b"].indexOf(seg.substring(1)),
                number: +seg.substring(0, 1)
            }
        })
        return {
            type: "fruit",
            repeat: 1,
            elements: fruits
        }
    }
}

let buffer = fs.readFileSync("./source.md")
let lines = buffer.toString().split("\n")
    .map(line => {
        let result = line.match(/!\[(.+).png]\((.+)\)/)
        return {
            image: result[2],
            ...nameToType(result[1])
        }
    });

let result = JSON.stringify(lines, null, 4)

console.log(result);
fs.writeFileSync("../src/cards.json", result)
