<div id="top"></div>



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/Azure/draftv2">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">DraftV2</h3>

  <p align="center">
    A tool to help developers hit the ground running with k8s
    <br />
    <a href="https://github.com/Azure/draftv2"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/Azure/draftv2">View Demo</a>
    ·
    <a href="https://github.com/Azure/draftv2/issues">Report Bug</a>
    ·
    <a href="https://github.com/Azure/draftv2/issues">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

[![Draft Screen Shot][product-screenshot]](https://example.com)

Draftv2 aims to simplify starting out with k8s. Draftv2 will create both a working Dockerfile for your application and create the necessary kubernetes manifests needed to hit the ground running with tools like Skaffold.
<p align="right">(<a href="#top">back to top</a>)</p>



### Built With

* [Go](https://go.dev/)
* [Draft](https://github.com/Azure/draft/)

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

Draftv2 requires Go version 1.17.x.
* Go
  ```sh
  go version
  ```

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/Azure/draftv2.git
   ```
2. Build the binary
   ```sh
   make
   ```
3. Add the binary to your path
   ```js
   mv draftv2 $GOPATH/bin/
   ```

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

Use this space to show useful examples of how a project can be used. Additional screenshots, code examples and demos work well in this space. You may also link to more resources.

_For more examples, please refer to the [Documentation](https://example.com)_

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [] Feature 1
- [] Feature 2
- [] Feature 3
    - [] Nested Feature

See the [open issues](https://github.com/Azure/draftv2/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Regenerate Integration Tests by running `./test/gen_integration.sh`
4. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
5. Push to the Branch (`git push origin feature/AmazingFeature`)
6. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

Your Name - [@twitter_handle](https://twitter.com/twitter_handle) - email@email_client.com

Project Link: [https://github.com/Azure/draftv2](https://github.com/Azure/draftv2)

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* []()
* []()
* []()

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/Azure/draftv2.svg?style=for-the-badge
[contributors-url]: https://github.com/Azure/draftv2/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/Azure/draftv2.svg?style=for-the-badge
[forks-url]: https://github.com/Azure/draftv2/network/members
[stars-shield]: https://img.shields.io/github/stars/Azure/draftv2.svg?style=for-the-badge
[stars-url]: https://github.com/Azure/draftv2/stargazers
[issues-shield]: https://img.shields.io/github/issues/Azure/draftv2.svg?style=for-the-badge
[issues-url]: https://github.com/Azure/draftv2/issues
[license-shield]: https://img.shields.io/github/license/Azure/draftv2.svg?style=for-the-badge
[license-url]: https://github.com/Azure/draftv2/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/linkedin_username
[product-screenshot]: images/screenshot.png