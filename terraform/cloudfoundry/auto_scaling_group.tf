// This launch template is required in the config
// but it should never cause any instances to be created.
resource "aws_launch_configuration" "routers_fake_launch_config" {
  name          = "${var.env}-router-launch-config"
  image_id      = "ami-0331f81234e646ca8"
  instance_type = "t3.micro"
}

// This autoscaling group will be used to syrchronize
// a number of load balancer target groups. Terraform
// should not be in charge of how many instances there
// are in the ASG.
resource "aws_autoscaling_group" "routers_asg" {
  name                 = "${var.env}-gorouter-asg"
  launch_configuration = "${aws_launch_configuration.routers_fake_launch_config.name}"

  vpc_zone_identifier = [
    "${aws_subnet.router.*.id}"
  ]

  min_size = 0
  max_size = 20

  tag {
    key = "deploy_env"
    value = "${var.env}"
    propagate_at_launch = false
  }
}
